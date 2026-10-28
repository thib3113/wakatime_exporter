/*
Copyright 2020 Jacob Colvin (MacroPower)
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collector

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/version"
)

// Auth modes supported by FetchHTTP.
const (
	AuthModeAuto   = "auto"
	AuthModeBasic  = "basic"
	AuthModeBearer = "bearer"
)

// OAuthTokenPrefix is the prefix identifying a WakaTime OAuth access token.
const OAuthTokenPrefix = "waka_tok_"

// userAgentProduct identifies the exporter in WakaTime request logs.
const userAgentProduct = "wakatime-exporter"

// repositoryURL is the public project repository.
const repositoryURL = "https://github.com/thib3113/wakatime_exporter"

// maxTokenLen caps every token embedded in the User-Agent, so neither the
// build version nor an inbound scraper value can bloat the header.
const maxTokenLen = 32

// formatUserAgent builds the RFC 9110 "product/version (comment)" form so
// requests are attributable to this exporter, its version and platform, and
// clearly identified as a Prometheus exporter rather than an IDE plugin.
// scraper, when non-empty, names the scraper that triggered the request; it
// must already be reduced by scraperToken.
func formatUserAgent(v string, scraper string) string {
	v = sanitizeToken(strings.TrimPrefix(v, "v"))
	if v == "" {
		v = "dev"
	}
	comment := "Prometheus exporter"
	if scraper != "" {
		comment += "; scraper=" + scraper
	}
	return fmt.Sprintf("%s/%s (%s; %s %s/%s; +%s)",
		userAgentProduct, v, comment, runtime.Version(), runtime.GOOS, runtime.GOARCH, repositoryURL)
}

// userAgent is the default User-Agent, resolved once: the build version is
// fixed at link time and the platform cannot change at runtime, so there is
// nothing to recompute per request.
var userAgent = formatUserAgent(version.Version, "")

// knownScrapers maps a lowercased User-Agent product to its canonical name.
// Anything else is dropped: inbound values are client-controlled and must
// never be forwarded verbatim to a third party.
var knownScrapers = map[string]string{
	"prometheus":      "Prometheus",
	"thanos":          "Thanos",
	"cortex":          "Cortex",
	"mimir":           "Mimir",
	"vmagent":         "vmagent",
	"vm_promscrape":   "vmagent",
	"grafanaagent":    "GrafanaAgent",
	"grafana-agent":   "GrafanaAgent",
	"alloy":           "Alloy",
	"victoriametrics": "VictoriaMetrics",
}

// scraperToken reduces a scraper User-Agent to a compact, allowlisted
// "Product/version" token, or "" when the scraper is not recognised. The
// version is capped and truncated at the first character outside
// [A-Za-z0-9._+-], so no arbitrary input (CRLF, spaces, quotes) can reach the
// outbound header.
func scraperToken(ua string) string {
	token := strings.TrimSpace(ua)
	if i := strings.IndexAny(token, " \t"); i >= 0 {
		token = token[:i]
	}
	if token == "" {
		return ""
	}
	name, ver, _ := strings.Cut(token, "/")
	canonical, known := knownScrapers[strings.ToLower(name)]
	if !known {
		return ""
	}
	ver = sanitizeToken(ver)
	if ver == "" {
		return canonical
	}
	return canonical + "/" + ver
}

// sanitizeToken keeps the leading run of [A-Za-z0-9._+-] characters, capped at
// maxTokenLen, and drops everything from the first invalid character onwards.
func sanitizeToken(token string) string {
	var b strings.Builder
	for _, r := range token {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '.', r == '_', r == '-', r == '+':
			if b.Len() == maxTokenLen {
				return b.String()
			}
			b.WriteRune(r)
		default:
			return b.String()
		}
	}
	return b.String()
}

type scrapeUserAgentKey struct{}

// WithScrapeUserAgent records the scraper User-Agent for the duration of one
// scrape, so outbound WakaTime requests can name the scraper that triggered
// them. It reports false when the User-Agent is empty or not recognised, in
// which case the context is returned unchanged.
func WithScrapeUserAgent(ctx context.Context, ua string) (context.Context, bool) {
	scraper := scraperToken(ua)
	if scraper == "" {
		return ctx, false
	}
	return context.WithValue(ctx, scrapeUserAgentKey{}, formatUserAgent(version.Version, scraper)), true
}

// userAgentFor returns the User-Agent to use for a request: the cached default
// unless the current scrape recorded an allowlisted scraper.
func userAgentFor(ctx context.Context) string {
	if ctx != nil {
		if ua, ok := ctx.Value(scrapeUserAgentKey{}).(string); ok && ua != "" {
			return ua
		}
	}
	return userAgent
}

// EffectiveAuthMode resolves the auth mode actually used for requests:
// an explicit "bearer"/"basic" mode wins, "auto" (or any unknown value)
// detects OAuth tokens by prefix. It never touches the token itself,
// so logging its result is safe.
func EffectiveAuthMode(token string, authMode string) string {
	switch authMode {
	case AuthModeBearer:
		return AuthModeBearer
	case AuthModeBasic:
		return AuthModeBasic
	default: // AuthModeAuto and any unknown value: detect OAuth tokens by prefix.
		if strings.HasPrefix(token, OAuthTokenPrefix) {
			return AuthModeBearer
		}
		return AuthModeBasic
	}
}

// AuthorizationHeader builds the value of the Authorization header for the
// given token and auth mode, without ever logging the token itself.
func AuthorizationHeader(token string, authMode string) string {
	if EffectiveAuthMode(token, authMode) == AuthModeBearer {
		return "Bearer " + token
	}
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(token))
}

// Namespace defines the common namespace to be used by all metrics.
const namespace = "wakatime"

const (
	maxConcurrentAPIRequests = 8
	maxResponseBytes         = 32 << 20
	apiSlotTimeout           = 30 * time.Second
)

var (
	apiRequestSlots = make(chan struct{}, maxConcurrentAPIRequests)
	rateLimitUntil  atomic.Int64
)

// FetchHTTP is a generic fetch method for Wakatime API endpoints
func FetchHTTP(client *http.Client, token string, authMode string, logger log.Logger) func(ctx context.Context, uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	authHeader := AuthorizationHeader(token, authMode)
	return func(ctx context.Context, uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		if ctx == nil {
			ctx = context.Background()
		}
		uri.Path = path.Join(uri.Path, subPath)
		uri.RawQuery = params.Encode()
		url := uri.String()

		level.Debug(logger).Log("msg", "Scraping Wakatime", "path", subPath, "url", uri.Redacted())

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", authHeader)
		req.Header.Set("User-Agent", userAgentFor(ctx))

		if until := rateLimitUntil.Load(); until > 0 && time.Now().Unix() < until {
			return nil, fmt.Errorf("rate limited by WakaTime until %s", time.Unix(until, 0).UTC().Format(time.RFC3339))
		}

		timer := time.NewTimer(apiSlotTimeout)
		select {
		case apiRequestSlots <- struct{}{}:
			timer.Stop()
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
			return nil, fmt.Errorf("timeout waiting for an API request slot")
		}
		resp, err := client.Do(req)
		if err != nil {
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
			<-apiRequestSlots
			return nil, err
		}

		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
			resp.Body.Close()
			<-apiRequestSlots
			return nil, httpStatusError(resp)
		}
		return &releasingReadCloser{
			ReadCloser: resp.Body,
			release:    func() { <-apiRequestSlots },
		}, nil
	}
}

func parseRetryAfter(value string) time.Duration {
	const defaultBackoff = 30 * time.Second
	if value == "" {
		return defaultBackoff
	}
	if seconds, err := strconv.Atoi(value); err == nil {
		return clampBackoff(time.Duration(seconds) * time.Second)
	}
	if when, err := http.ParseTime(value); err == nil {
		return clampBackoff(time.Until(when))
	}
	return defaultBackoff
}

func clampBackoff(delay time.Duration) time.Duration {
	if delay < time.Second {
		return time.Second
	}
	if delay > 5*time.Minute {
		return 5 * time.Minute
	}
	return delay
}

func httpStatusError(resp *http.Response) error {
	switch resp.StatusCode {
	case 401:
		return fmt.Errorf("HTTP status 401: unauthorized (check API key/token)")
	case 402:
		return fmt.Errorf("HTTP status 402: payment required (feature not available on this WakaTime plan)")
	case 403:
		return fmt.Errorf("HTTP status 403: forbidden (missing OAuth scope?)")
	case 429:
		delay := parseRetryAfter(resp.Header.Get("Retry-After"))
		rateLimitUntil.Store(time.Now().Add(delay).Unix())
		return fmt.Errorf("HTTP status 429: rate limited (retry after %s)", delay)
	default:
		return fmt.Errorf("HTTP status %d", resp.StatusCode)
	}
}

type releasingReadCloser struct {
	io.ReadCloser
	release func()
	once    sync.Once
}

func (r *releasingReadCloser) Close() error {
	err := r.ReadCloser.Close()
	r.once.Do(r.release)
	return err
}

// ReadAndUnmarshal reads the JSON response body and unmarshals the response
func ReadAndUnmarshal(body io.ReadCloser, object interface{}) error {
	respBody, readErr := io.ReadAll(io.LimitReader(body, maxResponseBytes+1))
	if readErr != nil {
		return readErr
	}
	if len(respBody) > maxResponseBytes {
		return fmt.Errorf("response exceeds %d bytes", maxResponseBytes)
	}

	if jsonErr := json.Unmarshal(respBody, object); jsonErr != nil {
		return jsonErr
	}

	return nil
}

/*
Copyright 2026 thib3113
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
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
)

func TestHTTPStatusError(t *testing.T) {
	defer rateLimitUntil.Store(0)
	cases := []struct {
		status int
		header http.Header
		want   string
	}{
		{401, http.Header{}, "unauthorized"},
		{402, http.Header{}, "payment required"},
		{403, http.Header{}, "forbidden"},
		{429, http.Header{"Retry-After": []string{"30"}}, "retry after 30"},
		{500, http.Header{}, "HTTP status 500"},
	}
	for _, tc := range cases {
		err := httpStatusError(&http.Response{StatusCode: tc.status, Header: tc.header})
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("status %d: got %v, want containing %q", tc.status, err, tc.want)
		}
	}
}

func TestFetchHTTPSetsUserAgentAndAuth(t *testing.T) {
	var gotUA, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		gotAuth = r.Header.Get("Authorization")
		_, _ = io.WriteString(w, `{}`)
	}))
	defer server.Close()

	uri, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse: %v", err)
	}

	body, err := FetchHTTP(server.Client(), "waka_tok_test", AuthModeAuto, log.NewNopLogger())(context.Background(), *uri, "stats/all_time", url.Values{})
	if err != nil {
		t.Fatalf("FetchHTTP: %v", err)
	}
	_ = body.Close()

	if gotUA != userAgent {
		t.Fatalf("User-Agent = %q, want %q", gotUA, userAgent)
	}
	if !strings.HasPrefix(gotUA, userAgentProduct+"/") {
		t.Fatalf("User-Agent = %q, want prefix %q", gotUA, userAgentProduct+"/")
	}
	if !strings.Contains(gotUA, "Prometheus exporter") {
		t.Fatalf("User-Agent = %q, want containing %q", gotUA, "Prometheus exporter")
	}
	if !strings.Contains(gotUA, "+"+repositoryURL) {
		t.Fatalf("User-Agent = %q, want containing %q", gotUA, "+"+repositoryURL)
	}
	if strings.Contains(gotUA, "scraper=") {
		t.Fatalf("User-Agent = %q, must not advertise a scraper without a scrape context", gotUA)
	}
	if gotAuth != "Bearer waka_tok_test" {
		t.Fatalf("Authorization = %q, want bearer", gotAuth)
	}
}

func TestFormatUserAgent(t *testing.T) {
	cases := []struct {
		name    string
		version string
		scraper string
		want    string
	}{
		{name: "empty falls back to dev", version: "", want: userAgentProduct + "/dev "},
		{name: "strips v prefix", version: "v1.2.3", want: userAgentProduct + "/1.2.3 "},
		{name: "keeps plain version", version: "1.2.3", want: userAgentProduct + "/1.2.3 "},
		{name: "adds scraper", version: "1.2.3", scraper: "Prometheus/2.53.0", want: userAgentProduct + "/1.2.3 "},
		{name: "sanitizes the build version", version: "1.2.3\r\nX-Evil: 1", want: userAgentProduct + "/1.2.3 "},
		{name: "invalid build version falls back to dev", version: "()", want: userAgentProduct + "/dev "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := formatUserAgent(tc.version, tc.scraper)
			if !strings.HasPrefix(got, tc.want) {
				t.Fatalf("formatUserAgent(%q, %q) = %q, want prefix %q", tc.version, tc.scraper, got, tc.want)
			}
			if strings.HasPrefix(got, userAgentProduct+"/v") {
				t.Fatalf("formatUserAgent(%q, %q) = %q, must not keep the v prefix", tc.version, tc.scraper, got)
			}
			if tc.scraper == "" && strings.Contains(got, "scraper=") {
				t.Fatalf("formatUserAgent(%q, \"\") = %q, must not advertise a scraper", tc.version, got)
			}
			if tc.scraper != "" && !strings.Contains(got, "scraper="+tc.scraper) {
				t.Fatalf("formatUserAgent(%q, %q) = %q, want scraper suffix", tc.version, tc.scraper, got)
			}
		})
	}
}

func TestScraperToken(t *testing.T) {
	cases := []struct {
		name string
		ua   string
		want string
	}{
		{name: "prometheus", ua: "Prometheus/2.53.0", want: "Prometheus/2.53.0"},
		{name: "lowercase product is canonicalised", ua: "prometheus/3.0.0", want: "Prometheus/3.0.0"},
		{name: "trailing comment is dropped", ua: "Prometheus/2.53.0 (linux/amd64)", want: "Prometheus/2.53.0"},
		{name: "versionless is kept", ua: "Alloy", want: "Alloy"},
		{name: "unknown product is dropped", ua: "curl/8.5.0", want: ""},
		{name: "browser is dropped", ua: "Mozilla/5.0 (X11; Linux)", want: ""},
		{name: "empty is dropped", ua: "", want: ""},
		{name: "crlf cannot be injected", ua: "Prometheus/2.53.0\r\nX-Evil: 1", want: "Prometheus/2.53.0"},
		{name: "nul cannot be injected", ua: "Prometheus/1.0\x00evil", want: "Prometheus/1.0"},
		{name: "quote stops the version", ua: `Prometheus/2.53.0"evil`, want: "Prometheus/2.53.0"},
		{name: "extra slash stops the version", ua: "Prometheus/1/2", want: "Prometheus/1"},
		{name: "long version is capped", ua: "Prometheus/" + strings.Repeat("9", 100), want: "Prometheus/" + strings.Repeat("9", maxTokenLen)},
		{name: "long product is dropped", ua: strings.Repeat("a", 100) + "/1.0", want: ""},
		{name: "surrounding spaces are trimmed", ua: " Prometheus /1.0", want: "Prometheus"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := scraperToken(tc.ua); got != tc.want {
				t.Fatalf("scraperToken(%q) = %q, want %q", tc.ua, got, tc.want)
			}
		})
	}
}

func TestFetchHTTPUsesScrapeUserAgent(t *testing.T) {
	var gotUA string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = io.WriteString(w, `{}`)
	}))
	defer server.Close()

	uri, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse: %v", err)
	}

	cases := []struct {
		name  string
		ua    string
		known bool
	}{
		{name: "known scraper", ua: "Prometheus/2.53.0", known: true},
		{name: "known scraper without version", ua: "Thanos", known: true},
		{name: "unknown scraper", ua: "curl/8.5.0", known: false},
		{name: "empty scraper", ua: "", known: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, known := WithScrapeUserAgent(context.Background(), tc.ua)
			if known != tc.known {
				t.Fatalf("WithScrapeUserAgent(%q) known = %v, want %v", tc.ua, known, tc.known)
			}
			body, err := FetchHTTP(server.Client(), "waka_tok_test", AuthModeAuto, log.NewNopLogger())(ctx, *uri, "stats/all_time", url.Values{})
			if err != nil {
				t.Fatalf("FetchHTTP: %v", err)
			}
			_ = body.Close()

			if !tc.known {
				if gotUA != userAgent {
					t.Fatalf("User-Agent = %q, want default %q", gotUA, userAgent)
				}
				return
			}
			token := scraperToken(tc.ua)
			if !strings.Contains(gotUA, "; scraper="+token+";") {
				t.Fatalf("User-Agent = %q, want scraper %q", gotUA, token)
			}
		})
	}
}

func TestParseRetryAfter(t *testing.T) {
	if got := parseRetryAfter("30"); got != 30*time.Second {
		t.Fatalf("parseRetryAfter(30) = %v, want 30s", got)
	}
	if got := parseRetryAfter(""); got != 30*time.Second {
		t.Fatalf("parseRetryAfter(empty) = %v, want 30s", got)
	}
	if got := parseRetryAfter("0"); got != time.Second {
		t.Fatalf("parseRetryAfter(0) = %v, want 1s", got)
	}
	if got := parseRetryAfter("999999"); got != 5*time.Minute {
		t.Fatalf("parseRetryAfter(big) = %v, want 5m", got)
	}
	if got := parseRetryAfter("garbage"); got != 30*time.Second {
		t.Fatalf("parseRetryAfter(garbage) = %v, want 30s", got)
	}
	if got := parseRetryAfter(time.Now().Add(10 * time.Second).UTC().Format(http.TimeFormat)); got < time.Second || got > 5*time.Minute {
		t.Fatalf("parseRetryAfter(date) = %v", got)
	}
}

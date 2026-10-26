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
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Auth modes supported by FetchHTTP.
const (
	AuthModeAuto   = "auto"
	AuthModeBasic  = "basic"
	AuthModeBearer = "bearer"
)

// OAuthTokenPrefix is the prefix identifying a WakaTime OAuth access token.
const OAuthTokenPrefix = "waka_tok_"

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
)

var apiRequestSlots = make(chan struct{}, maxConcurrentAPIRequests)

// BoolToBinary converts booleans to "0" or "1"
func BoolToBinary(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// FetchHTTP is a generic fetch method for Wakatime API endpoints
func FetchHTTP(token string, authMode string, sslVerify bool, timeout time.Duration, logger log.Logger) func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: !sslVerify}}
	client := http.Client{
		Timeout:   timeout,
		Transport: tr,
	}
	authHeader := AuthorizationHeader(token, authMode)
	return func(uri url.URL, subPath string, params url.Values) (io.ReadCloser, error) {
		uri.Path = path.Join(uri.Path, subPath)
		uri.RawQuery = params.Encode()
		url := uri.String()

		level.Debug(logger).Log("msg", "Scraping Wakatime", "path", subPath, "url", uri.Redacted())

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		req.Header = map[string][]string{
			"Authorization": {authHeader},
		}

		apiRequestSlots <- struct{}{}
		resp, err := client.Do(req)
		if err != nil {
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
			<-apiRequestSlots
			return nil, err
		}

		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			resp.Body.Close()
			<-apiRequestSlots
			return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
		}
		return &releasingReadCloser{
			ReadCloser: resp.Body,
			release:    func() { <-apiRequestSlots },
		}, nil
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

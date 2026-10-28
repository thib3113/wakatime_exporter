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

package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/thib3113/wakatime_exporter/collector"
)

func TestFilterKey(t *testing.T) {
	got := filterKey([]string{"stats", "summary", "stats", "all-time"})
	want := filterKey([]string{"all-time", "stats", "summary"})
	if got != want {
		t.Fatalf("filterKey = %q, want %q", got, want)
	}
}

func TestFilterKeyNoCollision(t *testing.T) {
	if filterKey([]string{"a,b"}) == filterKey([]string{"a", "b"}) {
		t.Fatal("filterKey must not collide on embedded separators")
	}
}

// TestScrapeUserAgentReachesUpstream exercises the whole chain: inbound scrape
// User-Agent -> scrape context -> collector -> outbound WakaTime request. It
// covers both the unfiltered and the filtered (collect[]) handler paths.
func TestScrapeUserAgentReachesUpstream(t *testing.T) {
	var mu sync.Mutex
	var upstream []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		upstream = append(upstream, r.Header.Get("User-Agent"))
		mu.Unlock()
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	defer server.Close()

	uri, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse: %v", err)
	}

	// Collector defaults are only applied by kingpin when flags are parsed.
	if _, err := kingpin.CommandLine.Parse([]string{"--collector.stats"}); err != nil {
		t.Fatalf("parse flags: %v", err)
	}

	h := &handler{
		exporterMetricsRegistry: prometheus.NewRegistry(),
		commonInputs: collector.CommonInputs{
			BaseURI:    *uri,
			URI:        *uri,
			Token:      "waka_tok_test",
			AuthMode:   "auto",
			Timeout:    5 * time.Second,
			HTTPClient: server.Client(),
			Config: collector.Config{Stats: collector.StatsConfig{
				Range:    "all_time",
				CacheTTL: time.Hour,
				Top:      1,
			}},
		},
		logger:           log.NewNopLogger(),
		filteredHandlers: map[string]http.Handler{},
	}
	h.unfilteredHandler, err = h.innerHandler("stats")
	if err != nil {
		t.Fatalf("innerHandler: %v", err)
	}

	for _, path := range []string{"/metrics", "/metrics?collect[]=stats"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("User-Agent", "Prometheus/2.53.0")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: status = %d, want 200", path, rec.Code)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(upstream) != 2 {
		t.Fatalf("upstream requests = %d, want 2", len(upstream))
	}
	for _, ua := range upstream {
		if !strings.Contains(ua, "; scraper=Prometheus/2.53.0;") {
			t.Fatalf("upstream User-Agent = %q, want scraper=Prometheus/2.53.0", ua)
		}
	}
}

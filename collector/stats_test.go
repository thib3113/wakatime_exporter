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
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

const statsTestJSON = `{
  "data": {
    "total_seconds": 432000,
    "total_seconds_including_other_language": 5461200,
    "daily_average": 3600,
    "daily_average_including_other_language": 3700,
    "ai_additions": 10,
    "ai_deletions": 2,
    "human_additions": 5,
    "human_deletions": 1,
    "ai_model_costs": {"model-alpha": 2.5},
    "ai_model_breakdown": [
      {"name": "model-alpha", "lines": 1234, "cost": 1.25},
      {"name": "model-beta", "lines": 567, "cost": 0.5}
    ],
    "ai_input_tokens": 2000,
    "ai_output_tokens": 800,
    "ai_prompt_events_total": 9,
    "ai_sessions": 50,
    "is_up_to_date": true,
    "range": "all_time",
    "best_day": {"date": "2026-01-15", "total_seconds": 36000},
    "days_including_holidays": 30,
    "days_minus_holidays": 20,
    "holidays": 10,
    "percent_calculated": 100,
    "is_including_today": true,
    "is_cached": true,
    "timeout": 15,
    "writes_only": false,
    "projects": [
      {
        "name": "example-project",
        "total_seconds": 432000,
        "ai_input_tokens": 1234,
        "ai_output_tokens": 567,
        "ai_prompt_events_total": 7,
        "ai_sessions": 42,
        "ai_model_costs": {"model-alpha": 1.25}
      },
      {
        "name": "small-project",
        "total_seconds": 60,
        "ai_input_tokens": 1,
        "ai_output_tokens": 1
      }
    ],
    "languages": [{"name": "TypeScript", "total_seconds": 4000000}],
    "editors": [{"name": "VS Code", "total_seconds": 5000000}],
    "categories": [{"name": "Coding", "total_seconds": 5000000}],
    "operating_systems": [{"name": "Linux", "total_seconds": 5000000}],
    "machines": [{"name": "laptop", "machine_name_id": "abc", "total_seconds": 5000000}],
    "dependencies": [{"name": "react", "total_seconds": 1000}]
  }
}`

// statsCollectorAdapter exposes a statsCollector to a prometheus.Registry,
// mirroring how WakaCollector drives UpdateContext with the scrape context.
// Production code never registers the collector directly.
type statsCollectorAdapter struct {
	c *statsCollector
}

func (a statsCollectorAdapter) Describe(ch chan<- *prometheus.Desc) {
	a.c.Describe(ch)
}

func (a statsCollectorAdapter) Collect(ch chan<- prometheus.Metric) {
	if err := a.c.UpdateContext(context.Background(), ch); err != nil && !isNoDataError(err) {
		panic(err)
	}
}

func TestStatsCollector(t *testing.T) {
	collector, err := NewStatsCollector(CommonInputs{
		Config: Config{Stats: StatsConfig{
			Range:    "all_time",
			CacheTTL: time.Hour,
			Top:      1,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewStatsCollector: %v", err)
	}

	calls := 0
	collector.(*statsCollector).fetchStat = func(context.Context, url.URL, string, url.Values) (io.ReadCloser, error) {
		calls++
		return io.NopCloser(strings.NewReader(statsTestJSON)), nil
	}

	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(statsCollectorAdapter{collector.(*statsCollector)})

	families, err := registry.Gather()
	if err != nil {
		t.Fatalf("Gather: %v", err)
	}

	assertGauge(t, families, "wakatime_stats_project_seconds", 432000,
		map[string]string{"range": "all_time", "name": "example-project"})
	assertGauge(t, families, "wakatime_stats_up_to_date", 1,
		map[string]string{"range": "all_time"})
	assertGauge(t, families, "wakatime_stats_project_ai_tokens", 1234,
		map[string]string{"range": "all_time", "name": "example-project", "kind": "input"})
	assertGauge(t, families, "wakatime_stats_project_ai_tokens", 567,
		map[string]string{"range": "all_time", "name": "example-project", "kind": "output"})
	assertGauge(t, families, "wakatime_stats_best_day_seconds", 36000,
		map[string]string{"range": "all_time", "date": "2026-01-15"})
	assertGauge(t, families, "wakatime_stats_days", 10,
		map[string]string{"range": "all_time", "kind": "holidays"})
	assertGauge(t, families, "wakatime_stats_ai_model_lines", 1234,
		map[string]string{"range": "all_time", "model": "model-alpha"})
	assertNoGauge(t, families, "wakatime_stats_ai_model_lines",
		map[string]string{"range": "all_time", "model": "model-beta"})
	assertNoGauge(t, families, "wakatime_stats_project_seconds",
		map[string]string{"range": "all_time", "name": "small-project"})

	if _, err := registry.Gather(); err != nil {
		t.Fatalf("second Gather: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected one API call thanks to the cache, got %d", calls)
	}
}

func TestStatsCollectorRejectsInvalidRange(t *testing.T) {
	if _, err := NewStatsCollector(CommonInputs{Config: Config{Stats: StatsConfig{Range: "../etc/passwd"}}}, log.NewNopLogger()); err == nil {
		t.Fatal("expected an error for an invalid range")
	}
}

func assertGauge(t *testing.T, families []*dto.MetricFamily, name string, want float64, labels map[string]string) {
	t.Helper()

	if metric := findMetric(families, name, labels); metric != nil {
		if got := metric.GetGauge().GetValue(); got != want {
			t.Fatalf("%s%v: got %v, want %v", name, labels, got, want)
		}
		return
	}
	t.Fatalf("metric %s%v not found", name, labels)
}

func assertNoGauge(t *testing.T, families []*dto.MetricFamily, name string, labels map[string]string) {
	t.Helper()

	if metric := findMetric(families, name, labels); metric != nil {
		t.Fatalf("metric %s%v should not be present", name, labels)
	}
}

func findMetric(families []*dto.MetricFamily, name string, labels map[string]string) *dto.Metric {
	for _, family := range families {
		if family.GetName() != name {
			continue
		}
		for _, metric := range family.GetMetric() {
			if hasLabels(metric, labels) {
				return metric
			}
		}
	}
	return nil
}

func hasLabels(metric *dto.Metric, want map[string]string) bool {
	for key, value := range want {
		found := false
		for _, label := range metric.GetLabel() {
			if label.GetName() == key && label.GetValue() == value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

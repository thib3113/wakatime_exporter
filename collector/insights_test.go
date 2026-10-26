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
	"testing"
	"time"

	"github.com/go-kit/kit/log"
)

func TestInsightsCollector(t *testing.T) {
	collector, err := NewInsightsCollector(CommonInputs{
		Config: Config{Insights: InsightsConfig{
			Types:    []string{"weekdays", "best_day", "daily_average", "ai_days"},
			Range:    "all_time",
			CacheTTL: time.Hour,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewInsightsCollector: %v", err)
	}
	collector.(*insightsCollector).fetchStat = fakeFetchByPath(map[string]string{
		"insights/weekdays/all_time":      `{"data":{"weekdays":[{"name":"Monday","total_seconds":3600},{"name":"Tuesday","total_seconds":1800}],"is_up_to_date":true}}`,
		"insights/best_day/all_time":      `{"data":{"best_day":{"date":"2026-01-15","total_seconds":7200},"is_up_to_date":true}}`,
		"insights/daily_average/all_time": `{"data":{"daily_average":{"total_seconds":1800},"is_up_to_date":true}}`,
		"insights/ai_days/all_time":       `{"data":{"ai_days":[{"date":"2026-01-15","total_seconds":3600,"ai_percent":42.5}],"is_up_to_date":true}}`,
	})

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_insight_weekday_seconds", 3600,
		map[string]string{"range": "all_time", "weekday": "Monday"})
	assertGaugeSample(t, samples, "wakatime_insight_best_day_seconds", 7200,
		map[string]string{"range": "all_time", "date": "2026-01-15"})
	assertGaugeSample(t, samples, "wakatime_insight_daily_average_seconds", 1800,
		map[string]string{"range": "all_time"})
	assertGaugeSample(t, samples, "wakatime_insight_day_seconds", 3600,
		map[string]string{"range": "all_time", "insight": "ai_days", "date": "2026-01-15"})
	assertGaugeSample(t, samples, "wakatime_insight_ai_days_ratio", 42.5,
		map[string]string{"range": "all_time", "date": "2026-01-15"})
	assertGaugeSample(t, samples, "wakatime_insight_up_to_date", 1,
		map[string]string{"range": "all_time", "insight": "weekdays"})
}

func TestInsightsCollectorRejectsInvalidType(t *testing.T) {
	if _, err := NewInsightsCollector(CommonInputs{
		Config: Config{Insights: InsightsConfig{Types: []string{"nope"}}},
	}, log.NewNopLogger()); err == nil {
		t.Fatal("expected an error for an invalid insight type")
	}
}

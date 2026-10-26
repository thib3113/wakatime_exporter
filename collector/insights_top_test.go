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

func TestInsightsCollectorTop(t *testing.T) {
	collector, err := NewInsightsCollector(CommonInputs{
		Config: Config{Insights: InsightsConfig{
			Types:      []string{"weekdays", "ai_days"},
			Range:      "all_time",
			CacheTTL:   time.Hour,
			Top:        1,
			MinSeconds: 0,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewInsightsCollector: %v", err)
	}
	collector.(*insightsCollector).fetchStat = fakeFetchByPath(map[string]string{
		"insights/weekdays/all_time": `{"data":{"weekdays":[{"name":"Monday","total_seconds":3600},{"name":"Tuesday","total_seconds":1800}],"is_up_to_date":true}}`,
		"insights/ai_days/all_time":  `{"data":{"ai_days":[{"date":"2026-01-15","total_seconds":3600,"ai_percent":42.5}],"is_up_to_date":true}}`,
	})

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_insight_weekday_seconds", 3600,
		map[string]string{"range": "all_time", "weekday": "Monday"})
	assertNoGaugeSample(t, samples, "wakatime_insight_weekday_seconds",
		map[string]string{"range": "all_time", "weekday": "Tuesday"})
	assertGaugeSample(t, samples, "wakatime_insight_ai_days_ratio", 42.5,
		map[string]string{"range": "all_time", "date": "2026-01-15"})
}

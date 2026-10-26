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

func TestStatsCollectorSkipsZeroAIMetrics(t *testing.T) {
	collector, err := NewStatsCollector(CommonInputs{
		Config: Config{Stats: StatsConfig{
			Range:      "all_time",
			CacheTTL:   time.Hour,
			Top:        0,
			MinSeconds: 0,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewStatsCollector: %v", err)
	}
	collector.(*statsCollector).fetchStat = fakeFetchJSON(`{
	  "data": {
	    "total_seconds": 100,
	    "is_up_to_date": true,
	    "projects": [
	      {"name": "ai-project", "total_seconds": 60, "ai_input_tokens": 10, "ai_output_tokens": 5},
	      {"name": "plain-project", "total_seconds": 40}
	    ]
	  }
	}`)

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_stats_project_ai_tokens", 10,
		map[string]string{"range": "all_time", "name": "ai-project", "kind": "input"})
	assertNoGaugeSample(t, samples, "wakatime_stats_project_ai_tokens",
		map[string]string{"range": "all_time", "name": "plain-project", "kind": "input"})
}

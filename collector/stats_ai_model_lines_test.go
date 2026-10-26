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

func TestStatsCollectorAIModelLinesFallback(t *testing.T) {
	collector, err := NewStatsCollector(CommonInputs{
		Config: Config{Stats: StatsConfig{
			Range:    "all_time",
			CacheTTL: time.Hour,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewStatsCollector: %v", err)
	}
	collector.(*statsCollector).fetchStat = fakeFetchJSON(`{
	  "data": {
	    "total_seconds": 100,
	    "is_up_to_date": true,
	    "ai_model_line_changes": {
	      "model-zeta": 50,
	      "model-alpha": 100,
	      "model-zero": 0
	    }
	  }
	}`)

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_stats_ai_model_lines", 100,
		map[string]string{"range": "all_time", "model": "model-alpha"})
	assertGaugeSample(t, samples, "wakatime_stats_ai_model_lines", 50,
		map[string]string{"range": "all_time", "model": "model-zeta"})
	assertNoGaugeSample(t, samples, "wakatime_stats_ai_model_lines",
		map[string]string{"range": "all_time", "model": "model-zero"})
}

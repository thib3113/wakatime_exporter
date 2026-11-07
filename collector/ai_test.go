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

	"github.com/go-kit/kit/log"
)

const aiTestJSON = `{
  "data": [
    {
      "project": "wakatime_exporter",
      "duration": 1800,
      "ai_input_tokens": 5000,
      "ai_output_tokens": 1200,
      "ai_cached_input_tokens": 3000,
      "ai_sessions": 2,
      "ai_prompt_events_total": 4,
      "ai_prompt_length_sum": 600,
      "ai_additions": 45,
      "ai_deletions": 10,
      "human_additions": 20,
      "human_deletions": 5,
      "ai_model_costs": {"DeepSeek": 0.05},
      "ai_model_breakdown": [
        {
          "name": "DeepSeek",
          "input_tokens": 5000,
          "output_tokens": 1200,
          "cached_input_tokens": 3000,
          "cost": 0.05,
          "lines": 45
        }
      ]
    }
  ]
}`

func TestAICollector(t *testing.T) {
	collector, err := NewAICollector(CommonInputs{
		Config: Config{AI: AIBreakdownConfig{
			Top:        10,
			MinSeconds: 0,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewAICollector: %v", err)
	}
	collector.(*aiCollector).fetchStat = fakeFetchJSON(aiTestJSON)

	samples := collectSamples(t, collector)
	assertCounterSample(t, samples, "wakatime_ai_tokens_total", 5000, map[string]string{"kind": "input"})
	assertCounterSample(t, samples, "wakatime_ai_tokens_total", 1200, map[string]string{"kind": "output"})
	assertCounterSample(t, samples, "wakatime_ai_tokens_total", 3000, map[string]string{"kind": "cached_input"})

	assertCounterSample(t, samples, "wakatime_ai_model_tokens_total", 5000, map[string]string{"model": "DeepSeek", "kind": "input"})
	assertCounterSample(t, samples, "wakatime_ai_model_tokens_total", 1200, map[string]string{"model": "DeepSeek", "kind": "output"})
	assertCounterSample(t, samples, "wakatime_ai_model_tokens_total", 3000, map[string]string{"model": "DeepSeek", "kind": "cached_input"})
	assertCounterSample(t, samples, "wakatime_ai_model_lines_total", 45, map[string]string{"model": "DeepSeek"})

	assertCounterSample(t, samples, "wakatime_ai_cost_usd_total", 0.05, map[string]string{"model": "DeepSeek"})
	assertCounterSample(t, samples, "wakatime_ai_sessions_total", 2, nil)
	assertCounterSample(t, samples, "wakatime_ai_prompt_events_total", 4, nil)
	assertGaugeSample(t, samples, "wakatime_ai_prompt_length_avg_chars", 150, nil)
	assertCounterSample(t, samples, "wakatime_durations_seconds_total", 1800, map[string]string{"name": "wakatime_exporter"})
}

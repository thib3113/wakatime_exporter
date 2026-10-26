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

const aiBreakdownTestJSON = `{
  "data": [
    {
      "language": "TypeScript",
      "duration": 3600,
      "ai_input_tokens": 1500,
      "ai_output_tokens": 500,
      "ai_sessions": 2,
      "ai_prompt_events_total": 3,
      "ai_prompt_length_sum": 120,
      "ai_additions": 10,
      "ai_deletions": 2,
      "human_additions": 5,
      "human_deletions": 1,
      "ai_model_costs": {"model-alpha": 0.5}
    },
    {
      "language": "Go",
      "duration": 600,
      "ai_input_tokens": 100,
      "ai_output_tokens": 50
    }
  ]
}`

func TestAIBreakdownCollector(t *testing.T) {
	collector, err := NewAIBreakdownCollector(CommonInputs{
		Config: Config{AI: AIBreakdownConfig{
			Dimensions: []string{"language"},
			Top:        10,
			MinTokens:  1000,
			CacheTTL:   time.Hour,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewAIBreakdownCollector: %v", err)
	}
	collector.(*aiBreakdownCollector).fetchStat = fakeFetchJSON(aiBreakdownTestJSON)

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_ai_breakdown_tokens", 1500,
		map[string]string{"slice": "language", "name": "TypeScript", "kind": "input"})
	assertGaugeSample(t, samples, "wakatime_ai_breakdown_tokens", 500,
		map[string]string{"slice": "language", "name": "TypeScript", "kind": "output"})
	assertGaugeSample(t, samples, "wakatime_ai_breakdown_cost_usd", 0.5,
		map[string]string{"slice": "language", "name": "TypeScript", "model": "model-alpha"})
	assertGaugeSample(t, samples, "wakatime_ai_breakdown_sessions", 2,
		map[string]string{"slice": "language", "name": "TypeScript"})
	assertNoGaugeSample(t, samples, "wakatime_ai_breakdown_tokens",
		map[string]string{"slice": "language", "name": "Go", "kind": "input"})
}

func TestAIBreakdownCollectorRejectsInvalidDimension(t *testing.T) {
	if _, err := NewAIBreakdownCollector(CommonInputs{
		Config: Config{AI: AIBreakdownConfig{Dimensions: []string{"not-a-dimension"}}},
	}, log.NewNopLogger()); err == nil {
		t.Fatal("expected an error for an invalid dimension")
	}
}

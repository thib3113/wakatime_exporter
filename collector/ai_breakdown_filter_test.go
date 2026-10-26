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

func TestAIBreakdownCollectorFilter(t *testing.T) {
	collector, err := NewAIBreakdownCollector(CommonInputs{
		Config: Config{AI: AIBreakdownConfig{
			Dimensions: []string{"language"},
			Top:        10,
			MinTokens:  0,
			CacheTTL:   time.Hour,
			Filter:     FilterConfig{Exclude: `^language:Go$`},
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewAIBreakdownCollector: %v", err)
	}
	collector.(*aiBreakdownCollector).fetchStat = fakeFetchJSON(aiBreakdownTestJSON)

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_ai_breakdown_tokens", 1500,
		map[string]string{"slice": "language", "name": "TypeScript", "kind": "input"})
	assertNoGaugeSample(t, samples, "wakatime_ai_breakdown_tokens",
		map[string]string{"slice": "language", "name": "Go", "kind": "input"})
}

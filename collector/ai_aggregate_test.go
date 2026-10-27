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

import "testing"

func TestAITotalsAddDuration(t *testing.T) {
	totals := newAITotals()
	totals.addDuration(wakatimeDurationEntry{
		AIAdditions:         10,
		AIDeletions:         2,
		HumanAdditions:      5,
		HumanDeletions:      1,
		AIInputTokens:       120,
		AIOutputTokens:      80,
		AISessions:          4,
		AIPromptEventsTotal: 6,
		AIPromptLengthSum:   300,
		AIModelCosts:        map[string]float64{"model-alpha": 0.5},
	})
	totals.addDuration(wakatimeDurationEntry{
		AIAdditions:         1,
		AIInputTokens:       20,
		AISessions:          2,
		AIPromptEventsTotal: 1,
		AIPromptLengthSum:   60,
		AIModelCosts:        map[string]float64{"model-alpha": 0.25, "model-beta": 0.1},
	})

	if totals.additions != 11 || totals.deletions != 2 || totals.humanAdditions != 5 || totals.humanDeletions != 1 {
		t.Fatalf("unexpected line totals: %#v", totals)
	}
	if totals.inputTokens != 140 || totals.outputTokens != 80 {
		t.Fatalf("unexpected token totals: %#v", totals)
	}
	if totals.sessions != 6 || totals.promptEvents != 7 || totals.promptLengthSum != 360 {
		t.Fatalf("unexpected AI totals: %#v", totals)
	}
	if totals.costs["model-alpha"] != 0.75 || totals.costs["model-beta"] != 0.1 {
		t.Fatalf("unexpected cost totals: %#v", totals.costs)
	}
}

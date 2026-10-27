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

type aiTotals struct {
	additions       int64
	deletions       int64
	humanAdditions  int64
	humanDeletions  int64
	inputTokens     int64
	outputTokens    int64
	sessions        int64
	promptEvents    int64
	promptLengthSum float64
	costs           map[string]float64
}

func newAITotals() *aiTotals {
	return &aiTotals{costs: map[string]float64{}}
}

func (a *aiTotals) addDuration(duration wakatimeDurationEntry) {
	a.additions += duration.AIAdditions
	a.deletions += duration.AIDeletions
	a.humanAdditions += duration.HumanAdditions
	a.humanDeletions += duration.HumanDeletions
	a.inputTokens += duration.AIInputTokens
	a.outputTokens += duration.AIOutputTokens
	a.sessions += duration.AISessions
	a.promptEvents += duration.AIPromptEventsTotal
	a.promptLengthSum += duration.AIPromptLengthSum
	for model, cost := range duration.AIModelCosts {
		a.costs[model] += cost
	}
}

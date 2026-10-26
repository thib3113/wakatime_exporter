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

// wakatimeDurations mirrors GET /users/{user}/durations?date=today.
// See https://wakatime.com/developers#durations
// Requires OAuth scope: read_heartbeats.
type wakatimeDurations struct {
	Data     []wakatimeDurationEntry `json:"data"`
	Start    string                  `json:"start"`
	End      string                  `json:"end"`
	Timezone string                  `json:"timezone"`
}

type wakatimeDurationEntry struct {
	Project         string  `json:"project"`
	Entity          string  `json:"entity"`
	Language        string  `json:"language"`
	Dependencies    string  `json:"dependencies"`
	OS              string  `json:"os"`
	OperatingSystem string  `json:"operating_system"`
	Editor          string  `json:"editor"`
	Category        string  `json:"category"`
	Machine         string  `json:"machine"`
	MachineName     string  `json:"machine_name"`
	MachineNameID   string  `json:"machine_name_id"`
	Time            float64 `json:"time"`
	Duration        float64 `json:"duration"`

	AIAdditions    int64 `json:"ai_additions"`
	AIDeletions    int64 `json:"ai_deletions"`
	HumanAdditions int64 `json:"human_additions"`
	HumanDeletions int64 `json:"human_deletions"`

	AIInputTokens  int64 `json:"ai_input_tokens"`
	AIOutputTokens int64 `json:"ai_output_tokens"`

	AIModelCosts map[string]float64 `json:"ai_model_costs"`

	AISessions                     int64   `json:"ai_sessions"`
	AIPromptEventsTotal            int64   `json:"ai_prompt_events_total"`
	AIPromptEventsAvgPerSession    float64 `json:"ai_prompt_events_avg_per_session"`
	AIPromptEventsMedianPerSession float64 `json:"ai_prompt_events_median_per_session"`

	AIPromptLengthAvg              float64 `json:"ai_prompt_length_avg"`
	AIPromptLengthAvgPerSession    float64 `json:"ai_prompt_length_avg_per_session"`
	AIPromptLengthMedianPerSession float64 `json:"ai_prompt_length_median_per_session"`
	AIPromptLengthSum              float64 `json:"ai_prompt_length_sum"`
}

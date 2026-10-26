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

type wakatimeHeartbeats struct {
	Data     flexibleList[wakatimeHeartbeat] `json:"data"`
	Start    string                          `json:"start"`
	End      string                          `json:"end"`
	Timezone string                          `json:"timezone"`
}

type wakatimeHeartbeat struct {
	Project        string  `json:"project"`
	Category       string  `json:"category"`
	MachineNameID  string  `json:"machine_name_id"`
	AISession      string  `json:"ai_session"`
	AIInputTokens  int64   `json:"ai_input_tokens"`
	AIOutputTokens int64   `json:"ai_output_tokens"`
	IsWrite        bool    `json:"is_write"`
	Time           float64 `json:"time"`
}

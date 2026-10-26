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

type wakatimeExternalDurations struct {
	Data     flexibleList[wakatimeExternalDuration] `json:"data"`
	Start    string                                 `json:"start"`
	End      string                                 `json:"end"`
	Timezone string                                 `json:"timezone"`
}

type wakatimeExternalDuration struct {
	ID         string  `json:"id"`
	ExternalID string  `json:"external_id"`
	Entity     string  `json:"entity"`
	Type       string  `json:"type"`
	Provider   string  `json:"provider"`
	Category   string  `json:"category"`
	StartTime  float64 `json:"start_time"`
	EndTime    float64 `json:"end_time"`
	Project    string  `json:"project"`
	Branch     string  `json:"branch"`
	Language   string  `json:"language"`
	Meta       string  `json:"meta"`
}

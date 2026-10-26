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

type wakatimeUserAgents struct {
	Data flexibleList[wakatimeUserAgent] `json:"data"`
}

type wakatimeUserAgent struct {
	ID                 string `json:"id"`
	Value              string `json:"value"`
	Editor             string `json:"editor"`
	AIModel            string `json:"ai_model"`
	AIModelVersion     string `json:"ai_model_version"`
	AIModelComplexity  string `json:"ai_model_complexity"`
	Version            string `json:"version"`
	OS                 string `json:"os"`
	LastSeenAt         string `json:"last_seen_at"`
	IsBrowserExtension bool   `json:"is_browser_extension"`
	IsDesktopApp       bool   `json:"is_desktop_app"`
	CreatedAt          string `json:"created_at"`
}

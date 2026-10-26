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

import "encoding/json"

type wakatimeInsights struct {
	Data       map[string]json.RawMessage `json:"data"`
	IsUpToDate bool                       `json:"is_up_to_date"`
}

type insightNamed struct {
	Name         string  `json:"name"`
	Day          string  `json:"day"`
	Weekday      string  `json:"weekday"`
	TotalSeconds float64 `json:"total_seconds"`
}

type insightDay struct {
	Date           string  `json:"date"`
	TotalSeconds   float64 `json:"total_seconds"`
	AIPercent      float64 `json:"ai_percent"`
	AILinesPercent float64 `json:"ai_lines_percent"`
	HumanPercent   float64 `json:"human_percent"`
	AILines        int64   `json:"ai_lines"`
	HumanLines     int64   `json:"human_lines"`
}

type insightBestDay struct {
	Date         string  `json:"date"`
	TotalSeconds float64 `json:"total_seconds"`
}

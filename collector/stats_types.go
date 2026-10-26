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

// wakatimeStats mirrors GET /users/{user}/stats/{range}.
// See https://wakatime.com/developers#stats
// Requires OAuth scope: read_stats.
type wakatimeStats struct {
	Data wakatimeStatsData `json:"data"`
}

type wakatimeStatsData struct {
	TotalSeconds                       float64 `json:"total_seconds"`
	TotalSecondsIncludingOtherLanguage float64 `json:"total_seconds_including_other_language"`
	DailyAverage                       float64 `json:"daily_average"`
	DailyAverageIncludingOtherLanguage float64 `json:"daily_average_including_other_language"`

	AIAdditions        int64                  `json:"ai_additions"`
	AIDeletions        int64                  `json:"ai_deletions"`
	HumanAdditions     int64                  `json:"human_additions"`
	HumanDeletions     int64                  `json:"human_deletions"`
	AIModelCosts       map[string]float64     `json:"ai_model_costs"`
	AIModelBreakdown   []wakatimeStatsAIModel `json:"ai_model_breakdown"`
	AIModelLineChanges map[string]int64       `json:"ai_model_line_changes"`
	AIInputTokens      int64                  `json:"ai_input_tokens"`
	AIOutputTokens     int64                  `json:"ai_output_tokens"`
	AIPromptEvents     int64                  `json:"ai_prompt_events_total"`
	AIPromptLengthSum  float64                `json:"ai_prompt_length_sum"`
	AISessions         int64                  `json:"ai_sessions"`

	Categories       []wakatimeStatsEntry   `json:"categories"`
	Dependencies     []wakatimeStatsEntry   `json:"dependencies"`
	Editors          []wakatimeStatsEntry   `json:"editors"`
	Languages        []wakatimeStatsEntry   `json:"languages"`
	Machines         []wakatimeStatsMachine `json:"machines"`
	OperatingSystems []wakatimeStatsEntry   `json:"operating_systems"`
	Projects         []wakatimeStatsProject `json:"projects"`

	BestDay               wakatimeStatsBestDay `json:"best_day"`
	DaysIncludingHolidays int                  `json:"days_including_holidays"`
	DaysMinusHolidays     int                  `json:"days_minus_holidays"`
	Holidays              int                  `json:"holidays"`
	PercentCalculated     int                  `json:"percent_calculated"`
	IsIncludingToday      bool                 `json:"is_including_today"`
	IsCached              bool                 `json:"is_cached"`
	Timeout               int                  `json:"timeout"`
	WritesOnly            bool                 `json:"writes_only"`
	HumanReadableRange    string               `json:"human_readable_range"`

	Range      string `json:"range"`
	IsUpToDate bool   `json:"is_up_to_date"`
}

type wakatimeStatsBestDay struct {
	Date         string  `json:"date"`
	TotalSeconds float64 `json:"total_seconds"`
}

type wakatimeStatsAIModel struct {
	Name  string  `json:"name"`
	Lines int64   `json:"lines"`
	Cost  float64 `json:"cost"`
}

type wakatimeStatsEntry struct {
	Name         string  `json:"name"`
	TotalSeconds float64 `json:"total_seconds"`
}

type wakatimeStatsMachine struct {
	Name          string  `json:"name"`
	MachineNameID string  `json:"machine_name_id"`
	TotalSeconds  float64 `json:"total_seconds"`
}

type wakatimeStatsProject struct {
	Name         string  `json:"name"`
	TotalSeconds float64 `json:"total_seconds"`

	AIAdditions    int64              `json:"ai_additions"`
	AIDeletions    int64              `json:"ai_deletions"`
	HumanAdditions int64              `json:"human_additions"`
	HumanDeletions int64              `json:"human_deletions"`
	AIModelCosts   map[string]float64 `json:"ai_model_costs"`
	AIInputTokens  int64              `json:"ai_input_tokens"`
	AIOutputTokens int64              `json:"ai_output_tokens"`
	AIPromptEvents int64              `json:"ai_prompt_events_total"`
	AISessions     int64              `json:"ai_sessions"`
}

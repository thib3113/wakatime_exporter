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

type wakatimePrivateLeaderboards struct {
	Data       flexibleList[wakatimePrivateLeaderboard] `json:"data"`
	Total      int                                      `json:"total"`
	TotalPages int                                      `json:"total_pages"`
}

type wakatimePrivateLeaderboard struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	TimeRange    string `json:"time_range"`
	MembersCount int    `json:"members_count"`
}

type wakatimePrivateLeaderboardLeaders struct {
	Data       flexibleList[wakatimePrivateLeaderboardLeader] `json:"data"`
	BoardType  string                                         `json:"board_type"`
	Page       int                                            `json:"page"`
	TotalPages int                                            `json:"total_pages"`
}

type wakatimePrivateLeaderboardLeader struct {
	Rank         int `json:"rank"`
	RunningTotal struct {
		AIModelTotalCost float64 `json:"ai_model_total_cost"`
	} `json:"running_total"`
	User struct {
		ID          string `json:"id"`
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
	} `json:"user"`
}

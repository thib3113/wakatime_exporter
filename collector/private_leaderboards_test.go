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

func TestPrivateLeaderboardsCollector(t *testing.T) {
	collector, err := NewPrivateLeaderboardsCollector(CommonInputs{
		Config: Config{PrivateLeaderboards: PrivateLeaderboardsConfig{
			Types:    []string{"ai"},
			MaxPages: 1,
			CacheTTL: time.Hour,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewPrivateLeaderboardsCollector: %v", err)
	}
	collector.(*privateLeaderboardsCollector).fetchStat = fakeFetchByPath(map[string]string{
		"leaderboards/b1": `{
		  "data": [
		    {"rank": 1, "running_total": {"ai_model_total_cost": 1.5}, "user": {"id": "u1", "username": "alice"}},
		    {"rank": 2, "running_total": {"ai_model_total_cost": 0.5}, "user": {"id": "u2", "username": "bob"}}
		  ],
		  "board_type": "ai",
		  "page": 1,
		  "total_pages": 1
		}`,
		"leaderboards": `{
		  "data": [
		    {"id": "b1", "name": "Friends", "time_range": "last_7_days", "members_count": 2}
		  ],
		  "total": 1,
		  "total_pages": 1
		}`,
	})

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_private_leaderboard_info", 1, map[string]string{
		"board":         "b1",
		"name":          "Friends",
		"time_range":    "last_7_days",
		"members_count": "2",
	})
	assertGaugeSample(t, samples, "wakatime_private_leaderboard_rank", 1, map[string]string{
		"board": "b1", "user": "alice", "board_type": "ai",
	})
	assertGaugeSample(t, samples, "wakatime_private_leaderboard_ai_spend_usd", 1.5, map[string]string{
		"board": "b1", "user": "alice", "board_type": "ai",
	})
	assertGaugeSample(t, samples, "wakatime_private_leaderboard_rank", 2, map[string]string{
		"board": "b1", "user": "bob", "board_type": "ai",
	})
	assertGaugeSample(t, samples, "wakatime_private_leaderboards_available", 1, nil)
	assertGaugeSample(t, samples, "wakatime_private_leaderboards_total", 1, nil)
}

func TestValidBoardID(t *testing.T) {
	for _, id := range []string{"b1", "123e4567-e89b-12d3-a456-426614174000", "board_1", "board.1"} {
		if !validBoardID(id) {
			t.Fatalf("%q should be valid", id)
		}
	}
	for _, id := range []string{"", ".", "..", "a/b", "a b", "../x"} {
		if validBoardID(id) {
			t.Fatalf("%q should be invalid", id)
		}
	}
}

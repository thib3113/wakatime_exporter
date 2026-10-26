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

func TestHeartbeatsCollector(t *testing.T) {
	collector, err := NewHeartbeatsCollector(CommonInputs{
		Config: Config{Heartbeats: HeartbeatsConfig{
			Top:      10,
			MinCount: 1,
			CacheTTL: time.Hour,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewHeartbeatsCollector: %v", err)
	}
	collector.(*heartbeatsCollector).fetchStat = fakeFetchJSON(`{
	  "data": [
	    {"project": "example-project", "is_write": true, "ai_input_tokens": 100, "ai_output_tokens": 50, "ai_session": "s1"},
	    {"project": "example-project", "is_write": false, "ai_input_tokens": 200, "ai_output_tokens": 60, "ai_session": "s1"},
	    {"project": "other-project", "is_write": false, "ai_input_tokens": 0, "ai_output_tokens": 0}
	  ]
	}`)

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_heartbeats", 2, map[string]string{"project": "example-project"})
	assertGaugeSample(t, samples, "wakatime_heartbeats_write", 1, map[string]string{"project": "example-project"})
	assertGaugeSample(t, samples, "wakatime_heartbeats_ai_tokens", 300,
		map[string]string{"project": "example-project", "kind": "input"})
	assertGaugeSample(t, samples, "wakatime_heartbeats_ai_tokens", 110,
		map[string]string{"project": "example-project", "kind": "output"})
	assertGaugeSample(t, samples, "wakatime_heartbeats_ai_sessions", 1, map[string]string{"project": "example-project"})
	assertGaugeSample(t, samples, "wakatime_heartbeats", 1, map[string]string{"project": "other-project"})
	assertGaugeSample(t, samples, "wakatime_heartbeats_total", 3, nil)
	assertGaugeSample(t, samples, "wakatime_heartbeats_exported_total", 3, nil)
}

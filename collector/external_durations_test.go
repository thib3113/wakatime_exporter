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

func TestExternalDurationsCollector(t *testing.T) {
	collector, err := NewExternalDurationsCollector(CommonInputs{
		Config: Config{ExternalDurations: ExternalDurationsConfig{
			Date:     "today",
			CacheTTL: time.Hour,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewExternalDurationsCollector: %v", err)
	}
	collector.(*externalDurationsCollector).fetchStat = fakeFetchJSON(`{
	  "data": [
	    {
	      "provider": "google_calendar",
	      "category": "communicating",
	      "project": "example-project",
	      "language": "",
	      "start_time": 1000,
	      "end_time": 1600
	    },
	    {
	      "provider": "google_calendar",
	      "category": "communicating",
	      "project": "example-project",
	      "language": "",
	      "start_time": 2000,
	      "end_time": 2300
	    }
	  ]
	}`)

	samples := collectSamples(t, collector)
	labels := map[string]string{
		"provider": "google_calendar",
		"category": "communicating",
		"project":  "example-project",
		"language": "",
	}
	assertGaugeSample(t, samples, "wakatime_external_seconds", 900, labels)
	assertGaugeSample(t, samples, "wakatime_external_durations", 2, labels)
	assertGaugeSample(t, samples, "wakatime_external_total_seconds", 900, nil)
}

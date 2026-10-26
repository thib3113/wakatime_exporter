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

func TestOrgsCollector(t *testing.T) {
	collector, err := NewOrgsCollector(CommonInputs{
		Config: Config{Orgs: OrgsConfig{CacheTTL: time.Hour}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewOrgsCollector: %v", err)
	}
	collector.(*orgsCollector).fetchStat = fakeFetchJSON(`{
	  "data": [
	    {
	      "id": "o1",
	      "name": "Acme",
	      "default_project_privacy": "visible",
	      "invited_people_count": 1,
	      "people_count": 10,
	      "timeout": 15,
	      "timezone": "Europe/Paris",
	      "writes_only": 0
	    }
	  ],
	  "total": 1,
	  "total_pages": 1
	}`)

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_org_info", 1, map[string]string{
		"id":                      "o1",
		"name":                    "Acme",
		"timezone":                "Europe/Paris",
		"default_project_privacy": "visible",
	})
	assertGaugeSample(t, samples, "wakatime_org_people_count", 10, map[string]string{"id": "o1", "name": "Acme"})
	assertGaugeSample(t, samples, "wakatime_org_invited_people_count", 1, map[string]string{"id": "o1", "name": "Acme"})
	assertGaugeSample(t, samples, "wakatime_org_timeout_minutes", 15, map[string]string{"id": "o1", "name": "Acme"})
	assertGaugeSample(t, samples, "wakatime_org_writes_only", 0, map[string]string{"id": "o1", "name": "Acme"})
	assertGaugeSample(t, samples, "wakatime_orgs_available", 1, nil)
	assertGaugeSample(t, samples, "wakatime_orgs_total", 1, nil)
}

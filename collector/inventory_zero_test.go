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

func TestMachinesCollectorAllFiltered(t *testing.T) {
	collector, err := NewMachinesCollector(CommonInputs{
		Config: Config{Inventory: InventoryConfig{
			Filter:   FilterConfig{Exclude: `^machine:.*$`},
			CacheTTL: time.Hour,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewMachinesCollector: %v", err)
	}
	collector.(*machinesCollector).fetchStat = fakeFetchJSON(`{
	  "data": [
	    {"id": "m1", "name": "old", "last_seen_at": "2026-02-01T12:00:00Z"},
	    {"id": "m2", "name": "new", "last_seen_at": "2026-02-02T12:00:00Z"}
	  ]
	}`)

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_machines_available", 2, nil)
	assertGaugeSample(t, samples, "wakatime_machines_total", 0, nil)
	assertNoGaugeSample(t, samples, "wakatime_machine_info", map[string]string{"id": "m1"})
}

func TestUserAgentsCollectorAllFiltered(t *testing.T) {
	collector, err := NewUserAgentsCollector(CommonInputs{
		Config: Config{Inventory: InventoryConfig{
			Filter:   FilterConfig{Exclude: `^editor:.*$`},
			CacheTTL: time.Hour,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewUserAgentsCollector: %v", err)
	}
	collector.(*userAgentsCollector).fetchStat = fakeFetchJSON(`{
	  "data": [
	    {"id": "ua1", "editor": "VS Code", "last_seen_at": "2026-02-01T12:00:00Z"},
	    {"id": "ua2", "editor": "Vim", "last_seen_at": "2026-02-02T12:00:00Z"}
	  ]
	}`)

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_user_agents_available", 2, nil)
	assertGaugeSample(t, samples, "wakatime_user_agents_total", 0, nil)
	assertNoGaugeSample(t, samples, "wakatime_user_agent_info", map[string]string{"id": "ua1"})
}

func TestProjectsCollectorAllFiltered(t *testing.T) {
	collector, err := NewProjectsCollector(CommonInputs{
		Config: Config{Inventory: InventoryConfig{
			Filter:   FilterConfig{Exclude: `^project:.*$`},
			CacheTTL: time.Hour,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewProjectsCollector: %v", err)
	}
	collector.(*projectsCollector).fetchStat = fakeFetchJSON(`{
	  "data": [
	    {"id": "1", "name": "example-project"},
	    {"id": "2", "name": "small-project"}
	  ]
	}`)

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_projects_available", 2, nil)
	assertGaugeSample(t, samples, "wakatime_projects_total", 0, nil)
	assertNoGaugeSample(t, samples, "wakatime_project_info",
		map[string]string{"name": "example-project"})
}

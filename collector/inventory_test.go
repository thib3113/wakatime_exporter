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

func TestProjectsCollector(t *testing.T) {
	collector, err := NewProjectsCollector(CommonInputs{
		Config: Config{Inventory: InventoryConfig{ProjectsFilter: "example", CacheTTL: time.Hour}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewProjectsCollector: %v", err)
	}
	collector.(*projectsCollector).fetchStat = fakeFetchJSON(`{
	  "data": [
	    {
	      "id": "1",
	      "name": "example-project",
	      "repository": "git@example.com:example.git",
	      "has_public_url": true,
	      "first_heartbeat_at": "2026-01-01T12:59:00Z",
	      "last_heartbeat_at": "2026-01-16T19:59:00Z",
	      "created_at": "2026-01-01T12:00:00Z"
	    },
	    {
	      "id": "2",
	      "name": "small-project",
	      "has_public_url": false,
	      "first_heartbeat_at": "",
	      "last_heartbeat_at": "",
	      "created_at": ""
	    }
	  ]
	}`)

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_project_info", 1, map[string]string{
		"name":       "example-project",
		"id":         "1",
		"repository": "git@example.com:example.git",
		"public_url": "true",
	})
	assertGaugeSample(t, samples, "wakatime_project_last_heartbeat_timestamp",
		float64(time.Date(2026, 1, 16, 19, 59, 0, 0, time.UTC).Unix()),
		map[string]string{"name": "example-project"})
	assertGaugeSample(t, samples, "wakatime_projects_available", 2, nil)
	assertGaugeSample(t, samples, "wakatime_projects_total", 1, nil)
	assertNoGaugeSample(t, samples, "wakatime_project_info", map[string]string{"name": "small-project"})
}

func TestUserAgentsCollector(t *testing.T) {
	collector, err := NewUserAgentsCollector(CommonInputs{
		Config: Config{Inventory: InventoryConfig{CacheTTL: time.Hour}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewUserAgentsCollector: %v", err)
	}
	collector.(*userAgentsCollector).fetchStat = fakeFetchJSON(`{
	  "data": [
	    {
	      "id": "ua1",
	      "editor": "VS Code",
	      "version": "1.2.3",
	      "os": "Linux",
	      "ai_model": "model-alpha",
	      "ai_model_version": "4",
	      "ai_model_complexity": "high",
	      "last_seen_at": "2026-02-01T12:00:00Z",
	      "is_browser_extension": false,
	      "is_desktop_app": true,
	      "created_at": "2026-01-01T00:00:00Z"
	    }
	  ]
	}`)

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_user_agent_info", 1, map[string]string{
		"id":                  "ua1",
		"editor":              "VS Code",
		"version":             "1.2.3",
		"os":                  "Linux",
		"ai_model":            "model-alpha",
		"ai_model_version":    "4",
		"ai_model_complexity": "high",
		"browser_extension":   "false",
		"desktop_app":         "true",
	})
	assertGaugeSample(t, samples, "wakatime_user_agent_last_seen_timestamp",
		float64(time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC).Unix()),
		map[string]string{"id": "ua1", "editor": "VS Code"})
	assertGaugeSample(t, samples, "wakatime_user_agents_available", 1, nil)
	assertGaugeSample(t, samples, "wakatime_user_agents_total", 1, nil)
}

func TestMachinesCollector(t *testing.T) {
	collector, err := NewMachinesCollector(CommonInputs{
		Config: Config{Inventory: InventoryConfig{CacheTTL: time.Hour}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewMachinesCollector: %v", err)
	}
	collector.(*machinesCollector).fetchStat = fakeFetchJSON(`{
	  "data": [
	    {
	      "id": "m1",
	      "name": "laptop",
	      "ip": "10.0.0.1",
	      "last_seen_at": "2026-02-01T12:00:00Z",
	      "timezone": "Europe/Paris",
	      "created_at": "2026-01-01T00:00:00Z"
	    }
	  ]
	}`)

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_machine_info", 1, map[string]string{
		"id":       "m1",
		"name":     "laptop",
		"ip":       "10.0.0.1",
		"timezone": "Europe/Paris",
	})
	assertGaugeSample(t, samples, "wakatime_machine_last_seen_timestamp",
		float64(time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC).Unix()),
		map[string]string{"id": "m1", "name": "laptop"})
	assertGaugeSample(t, samples, "wakatime_machines_available", 1, nil)
	assertGaugeSample(t, samples, "wakatime_machines_total", 1, nil)
}

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

func TestMachinesCollectorTop(t *testing.T) {
	collector, err := NewMachinesCollector(CommonInputs{
		Config: Config{Inventory: InventoryConfig{
			Top:      1,
			CacheTTL: time.Hour,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewMachinesCollector: %v", err)
	}
	collector.(*machinesCollector).fetchStat = fakeFetchJSON(`{
	  "data": [
	    {"id": "m1", "name": "old", "ip": "10.0.0.1", "last_seen_at": "2026-02-01T12:00:00Z", "created_at": "2026-01-01T00:00:00Z"},
	    {"id": "m2", "name": "new", "ip": "10.0.0.2", "last_seen_at": "2026-02-02T12:00:00Z", "created_at": "2026-01-01T00:00:00Z"}
	  ]
	}`)

	samples := collectSamples(t, collector)
	assertGaugeSample(t, samples, "wakatime_machine_info", 1,
		map[string]string{"id": "m2", "name": "new"})
	assertNoGaugeSample(t, samples, "wakatime_machine_info",
		map[string]string{"id": "m1", "name": "old"})
	assertGaugeSample(t, samples, "wakatime_machines_available", 2, nil)
	assertGaugeSample(t, samples, "wakatime_machines_total", 1, nil)
}

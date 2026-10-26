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

	"github.com/go-kit/kit/log"
)

func TestGoalCollectorEmptyChartData(t *testing.T) {
	collector, err := NewGoalCollector(CommonInputs{}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewGoalCollector: %v", err)
	}
	collector.(*goalCollector).fetchStat = fakeFetchJSON(`{
	  "data": [
	    {
	      "id": "goal-1",
	      "title": "Daily coding goal",
	      "type": "coding",
	      "delta": "day",
	      "is_enabled": true,
	      "chart_data": []
	    }
	  ],
	  "total": 1,
	  "total_pages": 1
	}`)

	if samples := collectSamples(t, collector); len(samples) != 0 {
		t.Fatalf("expected no metrics for an empty chart_data, got %d", len(samples))
	}
}

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
	"context"
	"sync"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

func TestStatsCollectorConcurrent(t *testing.T) {
	collector, err := NewStatsCollector(CommonInputs{
		Config: Config{Stats: StatsConfig{
			Range:    "all_time",
			CacheTTL: time.Hour,
		}},
	}, log.NewNopLogger())
	if err != nil {
		t.Fatalf("NewStatsCollector: %v", err)
	}
	collector.(*statsCollector).fetchStat = fakeFetchJSON(statsTestJSON)

	_ = collectSamples(t, collector)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := make(chan prometheus.Metric, 10000)
			if err := collector.UpdateContext(context.Background(), ch); err != nil {
				t.Errorf("Update: %v", err)
			}
			close(ch)
			for range ch {
			}
		}()
	}
	wg.Wait()
}

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
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

type contextCaptureCollector struct {
	got chan context.Context
}

func (c *contextCaptureCollector) UpdateContext(ctx context.Context, _ chan<- prometheus.Metric) error {
	c.got <- ctx
	return nil
}

func TestWakaCollectorPropagatesContext(t *testing.T) {
	got := make(chan context.Context, 1)
	collector := &WakaCollector{
		Collectors: map[string]Collector{"context": &contextCaptureCollector{got: got}},
		logger:     log.NewNopLogger(),
	}

	ctx, cancel := context.WithCancel(context.Background())
	collector.SetContext(ctx)
	ch := make(chan prometheus.Metric, 100)
	go func() {
		collector.Collect(ch)
		close(ch)
	}()

	select {
	case seen := <-got:
		if seen != ctx {
			t.Fatal("collector did not receive the scrape context")
		}
	case <-time.After(time.Second):
		t.Fatal("collector was not called")
	}
	cancel()
	for range ch {
	}
}

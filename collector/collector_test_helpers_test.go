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
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type metricSample struct {
	name   string
	metric *dto.Metric
}

func collectSamples(t *testing.T, c Collector) []metricSample {
	t.Helper()

	ch := make(chan prometheus.Metric, 10000)
	errCh := make(chan error, 1)
	go func() {
		errCh <- c.UpdateContext(context.Background(), ch)
		close(ch)
	}()

	samples := make([]metricSample, 0, 128)
	for metric := range ch {
		dtoMetric := &dto.Metric{}
		if err := metric.Write(dtoMetric); err != nil {
			t.Fatalf("write metric: %v", err)
		}
		samples = append(samples, metricSample{name: metricName(metric), metric: dtoMetric})
	}
	if err := <-errCh; err != nil {
		t.Fatalf("Update: %v", err)
	}
	return samples
}

func metricName(metric prometheus.Metric) string {
	const prefix = `Desc{fqName: "`
	desc := metric.Desc().String()
	if !strings.HasPrefix(desc, prefix) {
		return desc
	}
	rest := desc[len(prefix):]
	if index := strings.Index(rest, `"`); index >= 0 {
		return rest[:index]
	}
	return desc
}

func findSample(samples []metricSample, name string, labels map[string]string) *dto.Metric {
	for _, sample := range samples {
		if sample.name != name {
			continue
		}
		if hasLabels(sample.metric, labels) {
			return sample.metric
		}
	}
	return nil
}

func assertGaugeSample(t *testing.T, samples []metricSample, name string, want float64, labels map[string]string) {
	t.Helper()

	metric := findSample(samples, name, labels)
	if metric == nil {
		t.Fatalf("metric %s%v not found", name, labels)
	}
	if got := metric.GetGauge().GetValue(); got != want {
		t.Fatalf("%s%v: got %v, want %v", name, labels, got, want)
	}
}

func assertNoGaugeSample(t *testing.T, samples []metricSample, name string, labels map[string]string) {
	t.Helper()

	if metric := findSample(samples, name, labels); metric != nil {
		t.Fatalf("metric %s%v should not be present", name, labels)
	}
}

func fakeFetchJSON(body string) func(context.Context, url.URL, string, url.Values) (io.ReadCloser, error) {
	return func(context.Context, url.URL, string, url.Values) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(body)), nil
	}
}

func fakeFetchByPath(responses map[string]string) func(context.Context, url.URL, string, url.Values) (io.ReadCloser, error) {
	keys := make([]string, 0, len(responses))
	for key := range responses {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	return func(_ context.Context, _ url.URL, subPath string, _ url.Values) (io.ReadCloser, error) {
		for _, key := range keys {
			if strings.HasPrefix(subPath, key) {
				return io.NopCloser(strings.NewReader(responses[key])), nil
			}
		}
		return nil, fmt.Errorf("unexpected path %s", subPath)
	}
}

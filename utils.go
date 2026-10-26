/*
Copyright 2015 The Prometheus Authors
Modifications Copyright 2020 Jacob Colvin (MacroPower)
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

package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/thib3113/wakatime_exporter/collector"
)

const maxFilteredHandlers = 32

// handler wraps an unfiltered http.Handler but uses a filtered handler,
// created on the fly, if filtering is requested. Create instances with
// newHandler.
type handler struct {
	unfilteredHandler http.Handler
	// exporterMetricsRegistry is a separate registry for the metrics about
	// the exporter itself.
	exporterMetricsRegistry *prometheus.Registry
	includeExporterMetrics  bool
	commonInputs            collector.CommonInputs
	logger                  log.Logger

	mu               sync.Mutex
	filteredHandlers map[string]http.Handler
}

func newHandler(commonInputs collector.CommonInputs, includeExporterMetrics bool, logger log.Logger) *handler {
	h := &handler{
		exporterMetricsRegistry: prometheus.NewRegistry(),
		includeExporterMetrics:  includeExporterMetrics,
		commonInputs:            commonInputs,
		logger:                  logger,
		filteredHandlers:        map[string]http.Handler{},
	}
	if h.includeExporterMetrics {
		h.exporterMetricsRegistry.MustRegister(
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
			collectors.NewGoCollector(),
		)
	}
	if innerHandler, err := h.innerHandler(); err != nil {
		panic(fmt.Sprintf("Couldn't create metrics handler: %s", err))
	} else {
		h.unfilteredHandler = innerHandler
	}
	return h
}

// ServeHTTP implements http.Handler.
func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	filters := r.URL.Query()["collect[]"]
	level.Debug(h.logger).Log("msg", "collect query:", "filters", filters)

	if len(filters) == 0 {
		// No filters, use the prepared unfiltered handler.
		h.unfilteredHandler.ServeHTTP(w, r)
		return
	}
	// To serve filtered metrics, reuse the handler for this filter set so
	// collector caches (stats, insights, ...) survive across scrapes.
	key := filterKey(filters)
	h.mu.Lock()
	filteredHandler, ok := h.filteredHandlers[key]
	if !ok && len(h.filteredHandlers) >= maxFilteredHandlers {
		h.mu.Unlock()
		level.Warn(h.logger).Log("msg", "Too many collector filter combinations", "max", maxFilteredHandlers)
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte("too many collector filter combinations\n"))
		return
	}
	if !ok {
		var err error
		filteredHandler, err = h.innerHandler(filters...)
		if err != nil {
			h.mu.Unlock()
			level.Warn(h.logger).Log("msg", "Couldn't create filtered metrics handler:", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Couldn't create filtered metrics handler: %s", err)))
			return
		}
		h.filteredHandlers[key] = filteredHandler
	}
	h.mu.Unlock()

	filteredHandler.ServeHTTP(w, r)
}

func filterKey(filters []string) string {
	seen := make(map[string]struct{}, len(filters))
	unique := make([]string, 0, len(filters))
	for _, filter := range filters {
		if _, ok := seen[filter]; ok {
			continue
		}
		seen[filter] = struct{}{}
		unique = append(unique, filter)
	}
	sort.Strings(unique)
	var builder strings.Builder
	for _, filter := range unique {
		fmt.Fprintf(&builder, "%d:%s", len(filter), filter)
	}
	return builder.String()
}

// innerHandler is used to create both the one unfiltered http.Handler to be
// wrapped by the outer handler and also the filtered handlers created on the
// fly. The former is accomplished by calling innerHandler without any arguments
// (in which case it will log all the collectors enabled via command-line
// flags).
func (h *handler) innerHandler(filters ...string) (http.Handler, error) {
	nc, err := collector.NewWakaCollector(h.commonInputs, h.logger, filters...)
	if err != nil {
		return nil, fmt.Errorf("couldn't create collector: %s", err)
	}

	// Only log the creation of an unfiltered handler, which should happen
	// only once upon startup.
	if len(filters) == 0 {
		level.Info(h.logger).Log("msg", "Enabled collectors")
		collectors := []string{}
		for n := range nc.Collectors {
			collectors = append(collectors, n)
		}
		sort.Strings(collectors)
		for _, c := range collectors {
			level.Info(h.logger).Log("collector", c)
		}
	}

	r := prometheus.NewRegistry()
	if h.includeExporterMetrics {
		r.MustRegister(collectors.NewBuildInfoCollector())
	}
	if err := r.Register(nc); err != nil {
		return nil, fmt.Errorf("couldn't register collector: %s", err)
	}
	gatherers := prometheus.Gatherers{r}
	opts := promhttp.HandlerOpts{
		ErrorHandling: promhttp.ContinueOnError,
		Timeout:       30 * time.Second,
	}
	if h.includeExporterMetrics {
		gatherers = prometheus.Gatherers{h.exporterMetricsRegistry, r}
		opts.Registry = h.exporterMetricsRegistry
	}
	handler := promhttp.HandlerFor(gatherers, opts)
	if h.includeExporterMetrics {
		// Note that we have to use h.exporterMetricsRegistry here to
		// use the same promhttp metrics for all expositions.
		handler = promhttp.InstrumentMetricHandler(
			h.exporterMetricsRegistry, handler,
		)
	}
	return handler, nil
}

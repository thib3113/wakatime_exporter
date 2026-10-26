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
	"fmt"
	"io"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const aiBreakdownCollectorName = "ai-breakdown"

var aiBreakdownDimensions = map[string]bool{
	"project":      true,
	"entity":       true,
	"language":     true,
	"dependencies": true,
	"os":           true,
	"editor":       true,
	"category":     true,
	"machine":      true,
}

type aiBreakdownCollector struct {
	tokens          *prometheus.Desc
	costUSD         *prometheus.Desc
	lines           *prometheus.Desc
	sessions        *prometheus.Desc
	promptEvents    *prometheus.Desc
	promptLengthSum *prometheus.Desc
	dimensions      []string
	top             int
	minTokens       int64
	filter          *nameFilter
	cacheTTL        time.Duration
	uri             url.URL
	fetchStat       func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger          log.Logger

	mu      sync.Mutex
	cache   map[string]aiBreakdownCacheEntry
	flights flightGroup
}

type aiBreakdownCacheEntry struct {
	data  *wakatimeDurations
	until time.Time
}

type aiBreakdownAggregate struct {
	input           int64
	output          int64
	sessions        int64
	promptEvents    int64
	promptLengthSum float64
	aiAdditions     int64
	aiDeletions     int64
	humanAdditions  int64
	humanDeletions  int64
	costs           map[string]float64
}

func init() {
	registerCollector(aiBreakdownCollectorName, defaultEnabled, NewAIBreakdownCollector)
}

func NewAIBreakdownCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	dimensions := in.Config.AI.Dimensions
	if len(dimensions) == 0 {
		dimensions = []string{"language"}
	}
	for _, dimension := range dimensions {
		if !aiBreakdownDimensions[dimension] {
			return nil, fmt.Errorf("invalid AI breakdown dimension %q", dimension)
		}
	}

	cacheTTL := in.Config.AI.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = 15 * time.Minute
	}

	filter, err := newNameFilter(in.Config.AI.Filter)
	if err != nil {
		return nil, fmt.Errorf("invalid AI filter: %w", err)
	}

	return &aiBreakdownCollector{
		tokens: newMetricDesc(
			"ai_breakdown_tokens",
			"AI tokens used per breakdown dimension, from Wakatime durations.",
			"slice", "name", "kind",
		),
		costUSD: newMetricDesc(
			"ai_breakdown_cost_usd",
			"Estimated AI cost in USD according to WakaTime pricing per breakdown dimension.",
			"slice", "name", "model",
		),
		lines: newMetricDesc(
			"ai_breakdown_lines",
			"Lines attributed to AI vs human per breakdown dimension, from Wakatime durations.",
			"slice", "name", "origin", "kind",
		),
		sessions: newMetricDesc(
			"ai_breakdown_sessions",
			"Sum of per-duration AI sessions per breakdown dimension (may double-count overlapping sessions).",
			"slice", "name",
		),
		promptEvents: newMetricDesc(
			"ai_breakdown_prompt_events",
			"Prompts sent to AI tools per breakdown dimension, from Wakatime durations.",
			"slice", "name",
		),
		promptLengthSum: newMetricDesc(
			"ai_breakdown_prompt_length_chars_sum",
			"Sum of prompt characters per breakdown dimension, from Wakatime durations.",
			"slice", "name",
		),
		dimensions: dimensions,
		top:        in.Config.AI.Top,
		minTokens:  in.Config.AI.MinTokens,
		filter:     filter,
		cacheTTL:   cacheTTL,
		uri:        in.URI,
		fetchStat:  FetchHTTP(in.HTTPClient, in.Token, in.AuthMode, logger),
		logger:     logger,
		cache:      map[string]aiBreakdownCacheEntry{},
	}, nil
}

func (c *aiBreakdownCollector) Update(ch chan<- prometheus.Metric) error {
	emitted := false
	var firstErr error

	for _, dimension := range c.dimensions {
		data, err := c.getDurations(dimension)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			level.Error(c.logger).Log("msg", "AI breakdown fetch failed", "slice", dimension, "err", err)
			continue
		}
		if len(data.Data) == 0 {
			continue
		}
		c.emit(ch, dimension, data)
		emitted = true
	}

	if emitted {
		return nil
	}
	if firstErr != nil {
		return firstErr
	}
	return ErrNoData
}

func (c *aiBreakdownCollector) getDurations(dimension string) (*wakatimeDurations, error) {
	c.mu.Lock()
	if entry, ok := c.cache[dimension]; ok && time.Now().Before(entry.until) {
		data := entry.data
		c.mu.Unlock()
		return data, nil
	}
	c.mu.Unlock()

	return singleflightDo(&c.flights, dimension, func() (*wakatimeDurations, error) {
		params := url.Values{}
		params.Add("date", "today")
		params.Add("slice_by", dimension)
		params.Add("cache", "false")

		body, fetchErr := c.fetchStat(c.uri, aiEndpoint, params)
		if fetchErr != nil {
			return nil, fetchErr
		}
		defer body.Close()

		var response wakatimeDurations
		if err := ReadAndUnmarshal(body, &response); err != nil {
			return nil, err
		}

		c.mu.Lock()
		c.cache[dimension] = aiBreakdownCacheEntry{data: &response, until: time.Now().Add(c.cacheTTL)}
		c.mu.Unlock()
		return &response, nil
	})
}

func (c *aiBreakdownCollector) emit(ch chan<- prometheus.Metric, dimension string, data *wakatimeDurations) {
	byName := map[string]*aiBreakdownAggregate{}
	for _, duration := range data.Data {
		name := aiBreakdownValue(duration, dimension)
		if name == "" {
			name = "unknown"
		}
		aggregate, ok := byName[name]
		if !ok {
			aggregate = &aiBreakdownAggregate{costs: map[string]float64{}}
			byName[name] = aggregate
		}
		aggregate.input += duration.AIInputTokens
		aggregate.output += duration.AIOutputTokens
		aggregate.sessions += duration.AISessions
		aggregate.promptEvents += duration.AIPromptEventsTotal
		aggregate.promptLengthSum += duration.AIPromptLengthSum
		aggregate.aiAdditions += duration.AIAdditions
		aggregate.aiDeletions += duration.AIDeletions
		aggregate.humanAdditions += duration.HumanAdditions
		aggregate.humanDeletions += duration.HumanDeletions
		for model, cost := range duration.AIModelCosts {
			aggregate.costs[model] += cost
		}
	}

	type namedAggregate struct {
		name      string
		aggregate *aiBreakdownAggregate
	}
	entries := make([]namedAggregate, 0, len(byName))
	for name, aggregate := range byName {
		if !c.filter.Match(dimensionValue(dimension, name)) {
			continue
		}
		entries = append(entries, namedAggregate{name: name, aggregate: aggregate})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		left, right := aiBreakdownTotalTokens(entries[i].aggregate), aiBreakdownTotalTokens(entries[j].aggregate)
		if left == right {
			return entries[i].name < entries[j].name
		}
		return left > right
	})

	kept := entries[:0]
	for _, entry := range entries {
		if c.minTokens > 0 && aiBreakdownTotalTokens(entry.aggregate) < c.minTokens {
			continue
		}
		kept = append(kept, entry)
	}
	if c.top > 0 && len(kept) > c.top {
		kept = kept[:c.top]
	}

	level.Info(c.logger).Log(
		"msg", "Collecting AI breakdown from Wakatime",
		"slice", dimension,
		"durations", len(data.Data),
		"entries", len(kept),
	)

	for _, entry := range kept {
		name := entry.name
		aggregate := entry.aggregate
		ch <- prometheus.MustNewConstMetric(c.tokens, prometheus.GaugeValue, float64(aggregate.input), dimension, name, "input")
		ch <- prometheus.MustNewConstMetric(c.tokens, prometheus.GaugeValue, float64(aggregate.output), dimension, name, "output")
		for _, model := range sortedCostModels(aggregate.costs) {
			ch <- prometheus.MustNewConstMetric(c.costUSD, prometheus.GaugeValue, aggregate.costs[model], dimension, name, model)
		}
		ch <- prometheus.MustNewConstMetric(c.lines, prometheus.GaugeValue, float64(aggregate.aiAdditions), dimension, name, "ai", "additions")
		ch <- prometheus.MustNewConstMetric(c.lines, prometheus.GaugeValue, float64(aggregate.aiDeletions), dimension, name, "ai", "deletions")
		ch <- prometheus.MustNewConstMetric(c.lines, prometheus.GaugeValue, float64(aggregate.humanAdditions), dimension, name, "human", "additions")
		ch <- prometheus.MustNewConstMetric(c.lines, prometheus.GaugeValue, float64(aggregate.humanDeletions), dimension, name, "human", "deletions")
		ch <- prometheus.MustNewConstMetric(c.sessions, prometheus.GaugeValue, float64(aggregate.sessions), dimension, name)
		ch <- prometheus.MustNewConstMetric(c.promptEvents, prometheus.GaugeValue, float64(aggregate.promptEvents), dimension, name)
		ch <- prometheus.MustNewConstMetric(c.promptLengthSum, prometheus.GaugeValue, aggregate.promptLengthSum, dimension, name)
	}
}

func aiBreakdownValue(duration wakatimeDurationEntry, dimension string) string {
	switch dimension {
	case "project":
		return duration.Project
	case "entity":
		return duration.Entity
	case "language":
		return duration.Language
	case "dependencies":
		return duration.Dependencies
	case "os":
		if duration.OS != "" {
			return duration.OS
		}
		return duration.OperatingSystem
	case "editor":
		return duration.Editor
	case "category":
		return duration.Category
	case "machine":
		if duration.Machine != "" {
			return duration.Machine
		}
		if duration.MachineName != "" {
			return duration.MachineName
		}
		return duration.MachineNameID
	default:
		return ""
	}
}

func aiBreakdownTotalTokens(aggregate *aiBreakdownAggregate) int64 {
	return aggregate.input + aggregate.output
}

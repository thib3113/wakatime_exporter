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
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	heartbeatsCollectorName = "heartbeats"
	heartbeatsEndpoint      = "heartbeats"
)

type heartbeatsCollector struct {
	count         *prometheus.Desc
	writes        *prometheus.Desc
	aiTokens      *prometheus.Desc
	aiSessions    *prometheus.Desc
	total         *prometheus.Desc
	exportedTotal *prometheus.Desc
	top           int
	minCount      int
	filter        *nameFilter
	cache         ttlCache[wakatimeHeartbeats]
	cacheTTL      time.Duration
	uri           url.URL
	fetchStat     func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger        log.Logger
}

type heartbeatAggregate struct {
	count    int64
	writes   int64
	input    int64
	output   int64
	sessions map[string]struct{}
}

func init() {
	registerCollector(heartbeatsCollectorName, defaultDisabled, NewHeartbeatsCollector)
}

func NewHeartbeatsCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	cacheTTL := in.Config.Heartbeats.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = 15 * time.Minute
	}

	filter, err := newNameFilter(in.Config.Heartbeats.Filter)
	if err != nil {
		return nil, fmt.Errorf("invalid heartbeats filter: %w", err)
	}

	return &heartbeatsCollector{
		count: newMetricDesc(
			"heartbeats",
			"Number of heartbeats per project.",
			"project",
		),
		writes: newMetricDesc(
			"heartbeats_write",
			"Number of write heartbeats per project.",
			"project",
		),
		aiTokens: newMetricDesc(
			"heartbeats_ai_tokens",
			"AI tokens used in heartbeats per project.",
			"project", "kind",
		),
		aiSessions: newMetricDesc(
			"heartbeats_ai_sessions",
			"Distinct AI sessions seen in heartbeats per project.",
			"project",
		),
		total: newMetricDesc(
			"heartbeats_total",
			"Total number of heartbeats before filters and top limits.",
		),
		exportedTotal: newMetricDesc(
			"heartbeats_exported_total",
			"Number of heartbeats exported after filters and top limits.",
		),
		top:       in.Config.Heartbeats.Top,
		minCount:  in.Config.Heartbeats.MinCount,
		filter:    filter,
		cacheTTL:  cacheTTL,
		uri:       in.URI,
		fetchStat: FetchHTTP(in.HTTPClient, in.Token, in.AuthMode, logger),
		logger:    logger,
	}, nil
}

func (c *heartbeatsCollector) Update(ch chan<- prometheus.Metric) error {
	data, err := c.cache.get(c.cacheTTL, func() (*wakatimeHeartbeats, error) {
		params := url.Values{}
		params.Add("date", "today")

		body, fetchErr := c.fetchStat(c.uri, heartbeatsEndpoint, params)
		if fetchErr != nil {
			return nil, fetchErr
		}
		defer body.Close()

		var response wakatimeHeartbeats
		if err := ReadAndUnmarshal(body, &response); err != nil {
			return nil, err
		}
		return &response, nil
	})
	if err != nil {
		return err
	}
	if len(data.Data) == 0 {
		ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, 0)
		ch <- prometheus.MustNewConstMetric(c.exportedTotal, prometheus.GaugeValue, 0)
		return nil
	}

	byProject := map[string]*heartbeatAggregate{}
	for _, heartbeat := range data.Data {
		project := heartbeat.Project
		if project == "" {
			project = "unknown"
		}
		aggregate, ok := byProject[project]
		if !ok {
			aggregate = &heartbeatAggregate{sessions: map[string]struct{}{}}
			byProject[project] = aggregate
		}
		aggregate.count++
		if heartbeat.IsWrite {
			aggregate.writes++
		}
		aggregate.input += heartbeat.AIInputTokens
		aggregate.output += heartbeat.AIOutputTokens
		if heartbeat.AISession != "" {
			aggregate.sessions[heartbeat.AISession] = struct{}{}
		}
	}

	type projectAggregate struct {
		project   string
		aggregate *heartbeatAggregate
	}
	entries := make([]projectAggregate, 0, len(byProject))
	for project, aggregate := range byProject {
		entries = append(entries, projectAggregate{project: project, aggregate: aggregate})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].aggregate.count == entries[j].aggregate.count {
			return entries[i].project < entries[j].project
		}
		return entries[i].aggregate.count > entries[j].aggregate.count
	})

	kept := entries[:0]
	for _, entry := range entries {
		if !c.filter.Match(dimensionValue("project", entry.project)) {
			continue
		}
		if c.minCount > 0 && entry.aggregate.count < int64(c.minCount) {
			continue
		}
		kept = append(kept, entry)
	}
	if c.top > 0 && len(kept) > c.top {
		kept = kept[:c.top]
	}

	var exported int64
	for _, entry := range kept {
		exported += entry.aggregate.count
	}

	level.Info(c.logger).Log("msg", "Collecting heartbeats from Wakatime", "heartbeats", len(data.Data), "projects", len(kept))
	ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, float64(len(data.Data)))
	ch <- prometheus.MustNewConstMetric(c.exportedTotal, prometheus.GaugeValue, float64(exported))
	for _, entry := range kept {
		ch <- prometheus.MustNewConstMetric(c.count, prometheus.GaugeValue, float64(entry.aggregate.count), entry.project)
		ch <- prometheus.MustNewConstMetric(c.writes, prometheus.GaugeValue, float64(entry.aggregate.writes), entry.project)
		ch <- prometheus.MustNewConstMetric(c.aiTokens, prometheus.GaugeValue, float64(entry.aggregate.input), entry.project, "input")
		ch <- prometheus.MustNewConstMetric(c.aiTokens, prometheus.GaugeValue, float64(entry.aggregate.output), entry.project, "output")
		ch <- prometheus.MustNewConstMetric(c.aiSessions, prometheus.GaugeValue, float64(len(entry.aggregate.sessions)), entry.project)
	}

	return nil
}

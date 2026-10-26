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
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	userAgentsCollectorName = "user-agents"
	userAgentsEndpoint      = "user_agents"
)

type userAgentsCollector struct {
	info      *prometheus.Desc
	lastSeen  *prometheus.Desc
	created   *prometheus.Desc
	available *prometheus.Desc
	total     *prometheus.Desc
	filter    *nameFilter
	top       int
	cache     ttlCache[wakatimeUserAgents]
	cacheTTL  time.Duration
	uri       url.URL
	fetchStat func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger    log.Logger
}

func init() {
	registerCollector(userAgentsCollectorName, defaultEnabled, NewUserAgentsCollector)
}

func NewUserAgentsCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	cacheTTL := in.Config.Inventory.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = time.Hour
	}

	filter, err := newNameFilter(in.Config.Inventory.Filter)
	if err != nil {
		return nil, fmt.Errorf("invalid inventory filter: %w", err)
	}

	return &userAgentsCollector{
		info: newMetricDesc(
			"user_agent_info",
			"WakaTime user agents (IDE plugins) that have sent data.",
			"id", "editor", "version", "os", "ai_model", "ai_model_version", "ai_model_complexity", "browser_extension", "desktop_app",
		),
		lastSeen: newMetricDesc(
			"user_agent_last_seen_timestamp",
			"Unix timestamp when the user agent was last seen.",
			"id", "editor",
		),
		created: newMetricDesc(
			"user_agent_created_timestamp",
			"Unix timestamp when the user agent was first seen.",
			"id", "editor",
		),
		available: newMetricDesc(
			"user_agents_available",
			"Number of WakaTime user agents before filters and top limits.",
		),
		total: newMetricDesc(
			"user_agents_total",
			"Number of WakaTime user agents exported after filters and top limits.",
		),
		filter:    filter,
		top:       in.Config.Inventory.Top,
		cacheTTL:  cacheTTL,
		uri:       in.URI,
		fetchStat: FetchHTTP(in.HTTPClient, in.Token, in.AuthMode, logger),
		logger:    logger,
	}, nil
}

func (c *userAgentsCollector) Update(ch chan<- prometheus.Metric) error {
	data, err := c.cache.get(c.cacheTTL, func() (*wakatimeUserAgents, error) {
		body, fetchErr := c.fetchStat(c.uri, userAgentsEndpoint, url.Values{})
		if fetchErr != nil {
			return nil, fetchErr
		}
		defer body.Close()

		var response wakatimeUserAgents
		if err := ReadAndUnmarshal(body, &response); err != nil {
			return nil, err
		}
		return &response, nil
	})
	if err != nil {
		return err
	}
	if len(data.Data) == 0 {
		ch <- prometheus.MustNewConstMetric(c.available, prometheus.GaugeValue, 0)
		ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, 0)
		return nil
	}

	ch <- prometheus.MustNewConstMetric(c.available, prometheus.GaugeValue, float64(len(data.Data)))
	agents := selectInventory(data.Data, "editor",
		func(agent wakatimeUserAgent) string { return agent.Editor },
		func(agent wakatimeUserAgent) float64 { return parseUnixSecondsOrZero(agent.LastSeenAt) },
		c.filter, c.top)
	if len(agents) == 0 {
		ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, 0)
		return nil
	}

	level.Info(c.logger).Log("msg", "Collecting user agents from Wakatime", "count", len(agents))
	ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, float64(len(agents)))
	for _, agent := range agents {
		ch <- prometheus.MustNewConstMetric(c.info, prometheus.GaugeValue, 1,
			agent.ID,
			agent.Editor,
			agent.Version,
			agent.OS,
			agent.AIModel,
			agent.AIModelVersion,
			agent.AIModelComplexity,
			boolLabel(agent.IsBrowserExtension),
			boolLabel(agent.IsDesktopApp),
		)
		emitTimestamp(ch, c.lastSeen, agent.LastSeenAt, agent.ID, agent.Editor)
		emitTimestamp(ch, c.created, agent.CreatedAt, agent.ID, agent.Editor)
	}

	return nil
}

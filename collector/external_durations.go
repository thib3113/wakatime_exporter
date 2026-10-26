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
	externalDurationsCollectorName = "external-durations"
	externalDurationsEndpoint      = "external_durations"
)

type externalDurationsCollector struct {
	seconds   *prometheus.Desc
	count     *prometheus.Desc
	total     *prometheus.Desc
	date      string
	filter    *nameFilter
	cache     ttlCache[wakatimeExternalDurations]
	cacheTTL  time.Duration
	uri       url.URL
	fetchStat func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger    log.Logger
}

func validateExternalDurationsDate(value string) error {
	switch value {
	case "today", "yesterday":
		return nil
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return fmt.Errorf("invalid external durations date %q", value)
	}
	return nil
}

type externalDurationKey struct {
	provider string
	category string
	project  string
	language string
}

func init() {
	registerCollector(externalDurationsCollectorName, defaultEnabled, NewExternalDurationsCollector)
}

func NewExternalDurationsCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	date := in.Config.ExternalDurations.Date
	if date == "" {
		date = "today"
	}
	if err := validateExternalDurationsDate(date); err != nil {
		return nil, err
	}
	cacheTTL := in.Config.ExternalDurations.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = 15 * time.Minute
	}

	filter, err := newNameFilter(in.Config.ExternalDurations.Filter)
	if err != nil {
		return nil, fmt.Errorf("invalid external durations filter: %w", err)
	}

	return &externalDurationsCollector{
		seconds: newMetricDesc(
			"external_seconds",
			"Coding seconds from external durations (calendar, meetings, external apps).",
			"provider", "category", "project", "language",
		),
		count: newMetricDesc(
			"external_durations",
			"Number of external durations.",
			"provider", "category", "project", "language",
		),
		total: newMetricDesc(
			"external_total_seconds",
			"Total seconds from external durations after filters.",
		),
		date:      date,
		filter:    filter,
		cacheTTL:  cacheTTL,
		uri:       in.URI,
		fetchStat: FetchHTTP(in.HTTPClient, in.Token, in.AuthMode, logger),
		logger:    logger,
	}, nil
}

func (c *externalDurationsCollector) Update(ch chan<- prometheus.Metric) error {
	data, err := c.cache.get(c.cacheTTL, func() (*wakatimeExternalDurations, error) {
		params := url.Values{}
		params.Add("date", c.date)

		body, fetchErr := c.fetchStat(c.uri, externalDurationsEndpoint, params)
		if fetchErr != nil {
			return nil, fetchErr
		}
		defer body.Close()

		var response wakatimeExternalDurations
		if err := ReadAndUnmarshal(body, &response); err != nil {
			return nil, err
		}
		return &response, nil
	})
	if err != nil {
		return err
	}
	if len(data.Data) == 0 {
		return ErrNoData
	}

	seconds := map[externalDurationKey]float64{}
	counts := map[externalDurationKey]int{}
	var total float64
	for _, duration := range data.Data {
		length := duration.EndTime - duration.StartTime
		if length < 0 {
			continue
		}
		if !c.filter.MatchAny(
			dimensionValue("project", duration.Project),
			dimensionValue("provider", duration.Provider),
			dimensionValue("category", duration.Category),
			dimensionValue("language", duration.Language),
		) {
			continue
		}
		key := externalDurationKey{
			provider: duration.Provider,
			category: duration.Category,
			project:  duration.Project,
			language: duration.Language,
		}
		seconds[key] += length
		counts[key]++
		total += length
	}

	level.Info(c.logger).Log("msg", "Collecting external durations from Wakatime", "date", c.date, "count", len(data.Data))
	ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, total)
	for key, value := range seconds {
		ch <- prometheus.MustNewConstMetric(c.seconds, prometheus.GaugeValue, value, key.provider, key.category, key.project, key.language)
		ch <- prometheus.MustNewConstMetric(c.count, prometheus.GaugeValue, float64(counts[key]), key.provider, key.category, key.project, key.language)
	}

	return nil
}

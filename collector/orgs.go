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
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	orgsCollectorName = "orgs"
	orgsEndpoint      = "orgs"
	maxOrgPages       = 10
)

type orgsCollector struct {
	info         *prometheus.Desc
	peopleCount  *prometheus.Desc
	invitedCount *prometheus.Desc
	timeout      *prometheus.Desc
	writesOnly   *prometheus.Desc
	available    *prometheus.Desc
	total        *prometheus.Desc
	ids          []string
	cache        ttlCache[wakatimeOrgs]
	cacheTTL     time.Duration
	uri          url.URL
	fetchStat    func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger       log.Logger
}

func init() {
	registerCollector(orgsCollectorName, defaultDisabled, NewOrgsCollector)
}

func NewOrgsCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	cacheTTL := in.Config.Orgs.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = time.Hour
	}

	return &orgsCollector{
		info: newMetricDesc(
			"org_info",
			"WakaTime organization metadata.",
			"id", "name", "timezone", "default_project_privacy",
		),
		peopleCount: newMetricDesc(
			"org_people_count",
			"Number of people in the organization.",
			"id", "name",
		),
		invitedCount: newMetricDesc(
			"org_invited_people_count",
			"Number of pending invites in the organization.",
			"id", "name",
		),
		timeout: newMetricDesc(
			"org_timeout_minutes",
			"Keystroke timeout preference of the organization, in minutes.",
			"id", "name",
		),
		writesOnly: newMetricDesc(
			"org_writes_only",
			"Writes-only preference of the organization.",
			"id", "name",
		),
		available: newMetricDesc(
			"orgs_available",
			"Number of WakaTime organizations before filters.",
		),
		total: newMetricDesc(
			"orgs_total",
			"Number of WakaTime organizations exported after filters.",
		),
		ids:       in.Config.Orgs.IDs,
		cacheTTL:  cacheTTL,
		uri:       in.URI,
		fetchStat: FetchHTTP(in.HTTPClient, in.Token, in.AuthMode, logger),
		logger:    logger,
	}, nil
}

func (c *orgsCollector) Update(ch chan<- prometheus.Metric) error {
	data, err := c.cache.get(c.cacheTTL, func() (*wakatimeOrgs, error) {
		all := wakatimeOrgs{}
		for page := 1; page <= maxOrgPages; page++ {
			params := url.Values{}
			params.Set("page", strconv.Itoa(page))

			body, fetchErr := c.fetchStat(c.uri, orgsEndpoint, params)
			if fetchErr != nil {
				return nil, fetchErr
			}
			var response wakatimeOrgs
			unmarshalErr := ReadAndUnmarshal(body, &response)
			body.Close()
			if unmarshalErr != nil {
				return nil, unmarshalErr
			}

			all.Data = append(all.Data, response.Data...)
			all.Total = response.Total
			all.TotalPages = response.TotalPages
			if response.TotalPages == 0 || page >= response.TotalPages {
				break
			}
		}
		return &all, nil
	})
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(c.available, prometheus.GaugeValue, float64(len(data.Data)))

	allowed := map[string]bool{}
	for _, id := range c.ids {
		allowed[id] = true
	}

	emitted := 0
	for _, org := range data.Data {
		if len(allowed) > 0 && !allowed[org.ID] {
			continue
		}
		ch <- prometheus.MustNewConstMetric(c.info, prometheus.GaugeValue, 1,
			org.ID,
			org.Name,
			string(org.Timezone),
			org.DefaultProjectPrivacy,
		)
		ch <- prometheus.MustNewConstMetric(c.peopleCount, prometheus.GaugeValue, float64(org.PeopleCount), org.ID, org.Name)
		ch <- prometheus.MustNewConstMetric(c.invitedCount, prometheus.GaugeValue, float64(org.InvitedPeopleCount), org.ID, org.Name)
		ch <- prometheus.MustNewConstMetric(c.timeout, prometheus.GaugeValue, float64(org.Timeout), org.ID, org.Name)
		ch <- prometheus.MustNewConstMetric(c.writesOnly, prometheus.GaugeValue, float64(org.WritesOnly), org.ID, org.Name)
		emitted++
	}

	level.Info(c.logger).Log("msg", "Collecting organizations from Wakatime", "count", emitted)
	ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, float64(emitted))
	return nil
}

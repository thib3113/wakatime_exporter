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
	machinesCollectorName = "machines"
	machinesEndpoint      = "machine_names"
)

type machinesCollector struct {
	info      *prometheus.Desc
	lastSeen  *prometheus.Desc
	created   *prometheus.Desc
	available *prometheus.Desc
	total     *prometheus.Desc
	filter    *nameFilter
	top       int
	cache     ttlCache[wakatimeMachines]
	cacheTTL  time.Duration
	uri       url.URL
	fetchStat func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger    log.Logger
}

func init() {
	registerCollector(machinesCollectorName, defaultEnabled, NewMachinesCollector)
}

func NewMachinesCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	cacheTTL := in.Config.Inventory.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = time.Hour
	}

	filter, err := newNameFilter(in.Config.Inventory.Filter)
	if err != nil {
		return nil, fmt.Errorf("invalid inventory filter: %w", err)
	}

	return &machinesCollector{
		info: newMetricDesc(
			"machine_info",
			"WakaTime machine metadata.",
			"id", "name", "ip", "timezone",
		),
		lastSeen: newMetricDesc(
			"machine_last_seen_timestamp",
			"Unix timestamp when the machine was last seen.",
			"id", "name",
		),
		created: newMetricDesc(
			"machine_created_timestamp",
			"Unix timestamp when the machine was first seen.",
			"id", "name",
		),
		available: newMetricDesc(
			"machines_available",
			"Number of WakaTime machines before filters and top limits.",
		),
		total: newMetricDesc(
			"machines_total",
			"Number of WakaTime machines exported after filters and top limits.",
		),
		filter:    filter,
		top:       in.Config.Inventory.Top,
		cacheTTL:  cacheTTL,
		uri:       in.URI,
		fetchStat: FetchHTTP(in.HTTPClient, in.Token, in.AuthMode, logger),
		logger:    logger,
	}, nil
}

func (c *machinesCollector) Update(ch chan<- prometheus.Metric) error {
	data, err := c.cache.get(c.cacheTTL, func() (*wakatimeMachines, error) {
		body, fetchErr := c.fetchStat(c.uri, machinesEndpoint, url.Values{})
		if fetchErr != nil {
			return nil, fetchErr
		}
		defer body.Close()

		var response wakatimeMachines
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
	machines := make([]wakatimeMachine, 0, len(data.Data))
	for _, machine := range data.Data {
		if !c.filter.Match(dimensionValue("machine", machine.Name)) {
			continue
		}
		machines = append(machines, machine)
	}
	sort.SliceStable(machines, func(i, j int) bool {
		return parseUnixSeconds(machines[i].LastSeenAt) > parseUnixSeconds(machines[j].LastSeenAt)
	})
	if c.top > 0 && len(machines) > c.top {
		machines = machines[:c.top]
	}
	if len(machines) == 0 {
		ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, 0)
		return nil
	}

	level.Info(c.logger).Log("msg", "Collecting machines from Wakatime", "count", len(machines))
	ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, float64(len(machines)))
	for _, machine := range machines {
		ch <- prometheus.MustNewConstMetric(c.info, prometheus.GaugeValue, 1,
			machine.ID,
			machine.Name,
			machine.IP,
			machine.Timezone,
		)
		ch <- prometheus.MustNewConstMetric(c.lastSeen, prometheus.GaugeValue, parseUnixSeconds(machine.LastSeenAt), machine.ID, machine.Name)
		ch <- prometheus.MustNewConstMetric(c.created, prometheus.GaugeValue, parseUnixSeconds(machine.CreatedAt), machine.ID, machine.Name)
	}

	return nil
}

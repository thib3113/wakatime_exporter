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
	"regexp"
	"sort"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	projectsCollectorName = "projects"
	projectsEndpoint      = "projects"
)

type projectsCollector struct {
	info            *prometheus.Desc
	first           *prometheus.Desc
	last            *prometheus.Desc
	created         *prometheus.Desc
	available       *prometheus.Desc
	total           *prometheus.Desc
	projectFilter   *regexp.Regexp
	inventoryFilter *nameFilter
	top             int
	cache           ttlCache[wakatimeProjects]
	cacheTTL        time.Duration
	uri             url.URL
	fetchStat       func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger          log.Logger
}

func init() {
	registerCollector(projectsCollectorName, defaultEnabled, NewProjectsCollector)
}

func NewProjectsCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	var projectFilter *regexp.Regexp
	if in.Config.Inventory.ProjectsFilter != "" {
		compiled, err := regexp.Compile(in.Config.Inventory.ProjectsFilter)
		if err != nil {
			return nil, fmt.Errorf("invalid projects filter: %w", err)
		}
		projectFilter = compiled
	}

	inventoryFilter, err := newNameFilter(in.Config.Inventory.Filter)
	if err != nil {
		return nil, fmt.Errorf("invalid inventory filter: %w", err)
	}

	cacheTTL := in.Config.Inventory.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = time.Hour
	}

	return &projectsCollector{
		info: newMetricDesc(
			"project_info",
			"WakaTime project metadata.",
			"name", "id", "repository", "has_public_url",
		),
		first: newMetricDesc(
			"project_first_heartbeat_timestamp",
			"Unix timestamp of the first heartbeat for the project.",
			"name",
		),
		last: newMetricDesc(
			"project_last_heartbeat_timestamp",
			"Unix timestamp of the last heartbeat for the project.",
			"name",
		),
		created: newMetricDesc(
			"project_created_timestamp",
			"Unix timestamp when the project was created.",
			"name",
		),
		available: newMetricDesc(
			"projects_available",
			"Number of WakaTime projects before filters and top limits.",
		),
		total: newMetricDesc(
			"projects_total",
			"Number of WakaTime projects exported after filters and top limits.",
		),
		projectFilter:   projectFilter,
		inventoryFilter: inventoryFilter,
		top:             in.Config.Inventory.Top,
		cacheTTL:        cacheTTL,
		uri:             in.URI,
		fetchStat:       FetchHTTP(in.HTTPClient, in.Token, in.AuthMode, logger),
		logger:          logger,
	}, nil
}

func (c *projectsCollector) Update(ch chan<- prometheus.Metric) error {
	data, err := c.cache.get(c.cacheTTL, func() (*wakatimeProjects, error) {
		body, fetchErr := c.fetchStat(c.uri, projectsEndpoint, url.Values{})
		if fetchErr != nil {
			return nil, fetchErr
		}
		defer body.Close()

		var response wakatimeProjects
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
	projects := make([]wakatimeProject, 0, len(data.Data))
	for _, project := range data.Data {
		if c.projectFilter != nil && !c.projectFilter.MatchString(project.Name) {
			continue
		}
		if !c.inventoryFilter.Match(dimensionValue("project", project.Name)) {
			continue
		}
		projects = append(projects, project)
	}
	sort.SliceStable(projects, func(i, j int) bool {
		return parseUnixSeconds(projects[i].LastHeartbeatAt) > parseUnixSeconds(projects[j].LastHeartbeatAt)
	})
	if c.top > 0 && len(projects) > c.top {
		projects = projects[:c.top]
	}
	if len(projects) == 0 {
		ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, 0)
		return nil
	}

	for _, project := range projects {
		ch <- prometheus.MustNewConstMetric(c.info, prometheus.GaugeValue, 1,
			project.Name,
			project.ID,
			project.Repository,
			boolLabel(project.HasPublicURL),
		)
		ch <- prometheus.MustNewConstMetric(c.first, prometheus.GaugeValue, parseUnixSeconds(project.FirstHeartbeatAt), project.Name)
		ch <- prometheus.MustNewConstMetric(c.last, prometheus.GaugeValue, parseUnixSeconds(project.LastHeartbeatAt), project.Name)
		ch <- prometheus.MustNewConstMetric(c.created, prometheus.GaugeValue, parseUnixSeconds(project.CreatedAt), project.Name)
	}

	level.Info(c.logger).Log("msg", "Collecting projects from Wakatime", "count", len(projects))
	ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, float64(len(projects)))
	return nil
}

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

const (
	statsCollectorName   = "stats"
	statsSubsystem       = "stats"
	statsEndpoint        = "stats"
	defaultStatsRange    = "all_time"
	defaultStatsCacheTTL = time.Hour
)

type statsCollector struct {
	total                      *prometheus.Desc
	totalIncludingOther        *prometheus.Desc
	dailyAverage               *prometheus.Desc
	dailyAverageIncludingOther *prometheus.Desc
	bestDay                    *prometheus.Desc
	days                       *prometheus.Desc
	percentCalculated          *prometheus.Desc
	isIncludingToday           *prometheus.Desc
	isCached                   *prometheus.Desc
	timeoutMinutes             *prometheus.Desc
	writesOnly                 *prometheus.Desc
	project                    *prometheus.Desc
	language                   *prometheus.Desc
	editor                     *prometheus.Desc
	category                   *prometheus.Desc
	operatingSystem            *prometheus.Desc
	machine                    *prometheus.Desc
	dependency                 *prometheus.Desc
	aiTokens                   *prometheus.Desc
	aiCostUSD                  *prometheus.Desc
	aiLines                    *prometheus.Desc
	aiModelLines               *prometheus.Desc
	aiSessions                 *prometheus.Desc
	aiPromptEvents             *prometheus.Desc
	projectAITokens            *prometheus.Desc
	projectAICostUSD           *prometheus.Desc
	projectAILines             *prometheus.Desc
	projectAISessions          *prometheus.Desc
	projectAIPromptEvents      *prometheus.Desc
	upToDate                   *prometheus.Desc

	rangeName  string
	cacheTTL   time.Duration
	top        int
	minSeconds float64
	filter     *nameFilter
	uri        url.URL
	fetchStat  func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger     log.Logger

	mu          sync.Mutex
	cached      *wakatimeStatsData
	cachedUntil time.Time
	flights     flightGroup
}

func init() {
	registerCollector(statsCollectorName, defaultEnabled, NewStatsCollector)
}

func newStatsDesc(name, help string, labels ...string) *prometheus.Desc {
	return prometheus.NewDesc(prometheus.BuildFQName(namespace, statsSubsystem, name), help, labels, nil)
}

// NewStatsCollector returns a Collector exposing WakaTime stats for a range.
// Requires OAuth scope read_stats.
func NewStatsCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	rangeName := in.Config.Stats.Range
	if rangeName == "" {
		rangeName = defaultStatsRange
	}
	if err := validateRange(rangeName); err != nil {
		return nil, fmt.Errorf("invalid stats range: %w", err)
	}

	cacheTTL := in.Config.Stats.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = defaultStatsCacheTTL
	}

	filter, err := newNameFilter(in.Config.Stats.Filter)
	if err != nil {
		return nil, fmt.Errorf("invalid stats filter: %w", err)
	}

	return &statsCollector{
		total: newStatsDesc(
			"seconds",
			"Coding seconds for the configured range, excluding Other.",
			"range",
		),
		totalIncludingOther: newStatsDesc(
			"seconds_including_other",
			"Coding seconds for the configured range, including Other.",
			"range",
		),
		dailyAverage: newStatsDesc(
			"daily_average_seconds",
			"Average coding seconds per day for the configured range, excluding Other.",
			"range",
		),
		dailyAverageIncludingOther: newStatsDesc(
			"daily_average_seconds_including_other",
			"Average coding seconds per day for the configured range, including Other.",
			"range",
		),
		bestDay: newStatsDesc(
			"best_day_seconds",
			"Coding seconds on the best day for the configured range.",
			"range", "date",
		),
		days: newStatsDesc(
			"days",
			"Number of days in the configured range.",
			"range", "kind",
		),
		percentCalculated: newStatsDesc(
			"percent_calculated",
			"Percent of the stats background calculation completed.",
			"range",
		),
		isIncludingToday: newStatsDesc(
			"is_including_today",
			"Whether the stats include the current day.",
			"range",
		),
		isCached: newStatsDesc(
			"is_cached",
			"Whether the stats response came from the WakaTime cache.",
			"range",
		),
		timeoutMinutes: newStatsDesc(
			"timeout_minutes",
			"Keystroke timeout used for the stats, in minutes.",
			"range",
		),
		writesOnly: newStatsDesc(
			"writes_only",
			"Whether writes-only mode was used for the stats.",
			"range",
		),
		project: newStatsDesc(
			"project_seconds",
			"Coding seconds per project for the configured range.",
			"range", "name",
		),
		language: newStatsDesc(
			"language_seconds",
			"Coding seconds per language for the configured range.",
			"range", "name",
		),
		editor: newStatsDesc(
			"editor_seconds",
			"Coding seconds per editor for the configured range.",
			"range", "name",
		),
		category: newStatsDesc(
			"category_seconds",
			"Coding seconds per category for the configured range.",
			"range", "name",
		),
		operatingSystem: newStatsDesc(
			"operating_system_seconds",
			"Coding seconds per operating system for the configured range.",
			"range", "name",
		),
		machine: newStatsDesc(
			"machine_seconds",
			"Coding seconds per machine for the configured range.",
			"range", "id", "name",
		),
		dependency: newStatsDesc(
			"dependency_seconds",
			"Coding seconds per dependency for the configured range.",
			"range", "name",
		),
		aiTokens: newStatsDesc(
			"ai_tokens",
			"AI tokens used for the configured range.",
			"range", "kind",
		),
		aiCostUSD: newStatsDesc(
			"ai_cost_usd",
			"Estimated AI cost in USD according to WakaTime pricing for the configured range.",
			"range", "model",
		),
		aiLines: newStatsDesc(
			"ai_lines",
			"Lines attributed to AI vs human for the configured range.",
			"range", "origin", "kind",
		),
		aiModelLines: newStatsDesc(
			"ai_model_lines",
			"Lines added or removed by each AI model for the configured range.",
			"range", "model",
		),
		aiSessions: newStatsDesc(
			"ai_sessions",
			"Distinct AI sessions for the configured range.",
			"range",
		),
		aiPromptEvents: newStatsDesc(
			"ai_prompt_events",
			"Prompts sent to AI tools for the configured range.",
			"range",
		),
		projectAITokens: newStatsDesc(
			"project_ai_tokens",
			"AI tokens used per project for the configured range.",
			"range", "name", "kind",
		),
		projectAICostUSD: newStatsDesc(
			"project_ai_cost_usd",
			"Estimated AI cost in USD according to WakaTime pricing per project.",
			"range", "name", "model",
		),
		projectAILines: newStatsDesc(
			"project_ai_lines",
			"Lines attributed to AI vs human per project for the configured range.",
			"range", "name", "origin", "kind",
		),
		projectAISessions: newStatsDesc(
			"project_ai_sessions",
			"Distinct AI sessions per project for the configured range.",
			"range", "name",
		),
		projectAIPromptEvents: newStatsDesc(
			"project_ai_prompt_events",
			"Prompts sent to AI tools per project for the configured range.",
			"range", "name",
		),
		upToDate: newStatsDesc(
			"up_to_date",
			"Whether WakaTime reports the stats for the configured range as up to date.",
			"range",
		),
		rangeName:  rangeName,
		cacheTTL:   cacheTTL,
		top:        in.Config.Stats.Top,
		minSeconds: in.Config.Stats.MinSeconds,
		filter:     filter,
		uri:        in.URI,
		fetchStat:  FetchHTTP(in.HTTPClient, in.Token, in.AuthMode, logger),
		logger:     logger,
	}, nil
}

// Describe implements prometheus.Collector.
func (c *statsCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range []*prometheus.Desc{
		c.total,
		c.totalIncludingOther,
		c.dailyAverage,
		c.dailyAverageIncludingOther,
		c.bestDay,
		c.days,
		c.percentCalculated,
		c.isIncludingToday,
		c.isCached,
		c.timeoutMinutes,
		c.writesOnly,
		c.project,
		c.language,
		c.editor,
		c.category,
		c.operatingSystem,
		c.machine,
		c.dependency,
		c.aiTokens,
		c.aiCostUSD,
		c.aiLines,
		c.aiModelLines,
		c.aiSessions,
		c.aiPromptEvents,
		c.projectAITokens,
		c.projectAICostUSD,
		c.projectAILines,
		c.projectAISessions,
		c.projectAIPromptEvents,
		c.upToDate,
	} {
		ch <- desc
	}
}

// Collect implements prometheus.Collector.
func (c *statsCollector) Collect(ch chan<- prometheus.Metric) {
	if err := c.Update(ch); err != nil && !isNoDataError(err) {
		level.Error(c.logger).Log("msg", "stats collector failed", "err", err)
	}
}

func (c *statsCollector) Update(ch chan<- prometheus.Metric) error {
	data, err := c.getStats()
	if err != nil {
		return err
	}

	c.emit(ch, data)
	return nil
}

func (c *statsCollector) getStats() (*wakatimeStatsData, error) {
	c.mu.Lock()
	if c.cached != nil && time.Now().Before(c.cachedUntil) {
		cached := c.cached
		c.mu.Unlock()
		level.Debug(c.logger).Log("msg", "Using cached WakaTime stats", "range", c.rangeName)
		return cached, nil
	}
	c.mu.Unlock()

	return singleflightDo(&c.flights, "stats", func() (*wakatimeStatsData, error) {
		body, fetchErr := c.fetchStat(c.uri, statsEndpoint+"/"+c.rangeName, url.Values{})
		if fetchErr != nil {
			return nil, fetchErr
		}
		defer body.Close()

		var stats wakatimeStats
		if err := ReadAndUnmarshal(body, &stats); err != nil {
			return nil, err
		}

		data := &stats.Data
		level.Info(c.logger).Log(
			"msg", "Collecting stats from Wakatime",
			"range", c.rangeName,
			"up_to_date", data.IsUpToDate,
			"projects", len(data.Projects),
			"languages", len(data.Languages),
		)

		c.mu.Lock()
		c.cached = data
		if data.IsUpToDate {
			c.cachedUntil = time.Now().Add(c.cacheTTL)
		} else {
			c.cachedUntil = time.Now().Add(time.Minute)
		}
		c.mu.Unlock()

		return data, nil
	})
}

func (c *statsCollector) emit(ch chan<- prometheus.Metric, data *wakatimeStatsData) {
	r := c.rangeName
	metric := func(desc *prometheus.Desc, value float64, labels ...string) {
		ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value, labels...)
	}

	metric(c.upToDate, boolToFloat(data.IsUpToDate), r)
	metric(c.total, data.TotalSeconds, r)
	metric(c.totalIncludingOther, data.TotalSecondsIncludingOtherLanguage, r)
	metric(c.dailyAverage, data.DailyAverage, r)
	metric(c.dailyAverageIncludingOther, data.DailyAverageIncludingOtherLanguage, r)
	if data.BestDay.Date != "" {
		metric(c.bestDay, data.BestDay.TotalSeconds, r, data.BestDay.Date)
	}
	metric(c.days, float64(data.DaysIncludingHolidays), r, "including_holidays")
	metric(c.days, float64(data.DaysMinusHolidays), r, "minus_holidays")
	metric(c.days, float64(data.Holidays), r, "holidays")
	metric(c.percentCalculated, float64(data.PercentCalculated), r)
	metric(c.isIncludingToday, boolToFloat(data.IsIncludingToday), r)
	metric(c.isCached, boolToFloat(data.IsCached), r)
	metric(c.timeoutMinutes, float64(data.Timeout), r)
	metric(c.writesOnly, boolToFloat(data.WritesOnly), r)

	projects := trimProjects(data.Projects, c.top, c.minSeconds, c.filter)
	for _, entry := range projects {
		metric(c.project, entry.TotalSeconds, r, entry.Name)
		if entry.AIInputTokens != 0 || entry.AIOutputTokens != 0 ||
			entry.AISessions != 0 || entry.AIPromptEvents != 0 ||
			len(entry.AIModelCosts) != 0 ||
			entry.AIAdditions != 0 || entry.AIDeletions != 0 ||
			entry.HumanAdditions != 0 || entry.HumanDeletions != 0 {
			metric(c.projectAITokens, float64(entry.AIInputTokens), r, entry.Name, "input")
			metric(c.projectAITokens, float64(entry.AIOutputTokens), r, entry.Name, "output")
			for _, model := range sortedCostModels(entry.AIModelCosts) {
				metric(c.projectAICostUSD, entry.AIModelCosts[model], r, entry.Name, model)
			}
			metric(c.projectAILines, float64(entry.AIAdditions), r, entry.Name, "ai", "additions")
			metric(c.projectAILines, float64(entry.AIDeletions), r, entry.Name, "ai", "deletions")
			metric(c.projectAILines, float64(entry.HumanAdditions), r, entry.Name, "human", "additions")
			metric(c.projectAILines, float64(entry.HumanDeletions), r, entry.Name, "human", "deletions")
			metric(c.projectAISessions, float64(entry.AISessions), r, entry.Name)
			metric(c.projectAIPromptEvents, float64(entry.AIPromptEvents), r, entry.Name)
		}
	}
	for _, entry := range trimEntries(data.Languages, "language", c.top, c.minSeconds, c.filter) {
		metric(c.language, entry.TotalSeconds, r, entry.Name)
	}
	for _, entry := range trimEntries(data.Editors, "editor", c.top, c.minSeconds, c.filter) {
		metric(c.editor, entry.TotalSeconds, r, entry.Name)
	}
	for _, entry := range trimEntries(data.Categories, "category", c.top, c.minSeconds, c.filter) {
		metric(c.category, entry.TotalSeconds, r, entry.Name)
	}
	for _, entry := range trimEntries(data.OperatingSystems, "os", c.top, c.minSeconds, c.filter) {
		metric(c.operatingSystem, entry.TotalSeconds, r, entry.Name)
	}
	for _, entry := range trimMachines(data.Machines, c.top, c.minSeconds, c.filter) {
		metric(c.machine, entry.TotalSeconds, r, entry.MachineNameID, entry.Name)
	}
	for _, entry := range trimEntries(data.Dependencies, "dependency", c.top, c.minSeconds, c.filter) {
		metric(c.dependency, entry.TotalSeconds, r, entry.Name)
	}

	metric(c.aiTokens, float64(data.AIInputTokens), r, "input")
	metric(c.aiTokens, float64(data.AIOutputTokens), r, "output")
	for _, model := range sortedCostModels(data.AIModelCosts) {
		metric(c.aiCostUSD, data.AIModelCosts[model], r, model)
	}
	metric(c.aiLines, float64(data.AIAdditions), r, "ai", "additions")
	metric(c.aiLines, float64(data.AIDeletions), r, "ai", "deletions")
	metric(c.aiLines, float64(data.HumanAdditions), r, "human", "additions")
	metric(c.aiLines, float64(data.HumanDeletions), r, "human", "deletions")

	models := append([]wakatimeStatsAIModel(nil), data.AIModelBreakdown...)
	if len(models) == 0 {
		models = make([]wakatimeStatsAIModel, 0, len(data.AIModelLineChanges))
		for name, lines := range data.AIModelLineChanges {
			models = append(models, wakatimeStatsAIModel{Name: name, Lines: lines})
		}
	}
	sort.SliceStable(models, func(i, j int) bool {
		return models[i].Name < models[j].Name
	})
	for _, model := range models {
		if model.Name == "" || model.Lines == 0 {
			continue
		}
		metric(c.aiModelLines, float64(model.Lines), r, model.Name)
	}

	metric(c.aiSessions, float64(data.AISessions), r)
	metric(c.aiPromptEvents, float64(data.AIPromptEvents), r)
}

func boolToFloat(value bool) float64 {
	if value {
		return 1
	}
	return 0
}

func sortedCostModels(costs map[string]float64) []string {
	models := make([]string, 0, len(costs))
	for model := range costs {
		models = append(models, model)
	}
	sort.Strings(models)
	return models
}

func trimEntries(entries []wakatimeStatsEntry, dimension string, top int, minSeconds float64, filter *nameFilter) []wakatimeStatsEntry {
	out := make([]wakatimeStatsEntry, 0, len(entries))
	for _, entry := range entries {
		if !filter.Match(dimensionValue(dimension, entry.Name)) {
			continue
		}
		out = append(out, entry)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].TotalSeconds > out[j].TotalSeconds
	})
	if minSeconds > 0 {
		filtered := out[:0]
		for _, entry := range out {
			if entry.TotalSeconds >= minSeconds {
				filtered = append(filtered, entry)
			}
		}
		out = filtered
	}
	if top > 0 && len(out) > top {
		out = out[:top]
	}
	return out
}

func trimProjects(projects []wakatimeStatsProject, top int, minSeconds float64, filter *nameFilter) []wakatimeStatsProject {
	out := make([]wakatimeStatsProject, 0, len(projects))
	for _, project := range projects {
		if !filter.Match(dimensionValue("project", project.Name)) {
			continue
		}
		out = append(out, project)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].TotalSeconds > out[j].TotalSeconds
	})
	if minSeconds > 0 {
		filtered := out[:0]
		for _, project := range out {
			if project.TotalSeconds >= minSeconds {
				filtered = append(filtered, project)
			}
		}
		out = filtered
	}
	if top > 0 && len(out) > top {
		out = out[:top]
	}
	return out
}

func trimMachines(machines []wakatimeStatsMachine, top int, minSeconds float64, filter *nameFilter) []wakatimeStatsMachine {
	out := make([]wakatimeStatsMachine, 0, len(machines))
	for _, machine := range machines {
		if !filter.Match(dimensionValue("machine", machine.Name)) {
			continue
		}
		out = append(out, machine)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].TotalSeconds > out[j].TotalSeconds
	})
	if minSeconds > 0 {
		filtered := out[:0]
		for _, machine := range out {
			if machine.TotalSeconds >= minSeconds {
				filtered = append(filtered, machine)
			}
		}
		out = filtered
	}
	if top > 0 && len(out) > top {
		out = out[:top]
	}
	return out
}

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
	"encoding/json"
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
	insightsCollectorName = "insights"
	insightsEndpoint      = "insights"
	defaultInsightsRange  = "all_time"
)

var insightTypes = map[string]bool{
	"stats":             true,
	"weekdays":          true,
	"days":              true,
	"ai_days":           true,
	"best_day":          true,
	"daily_average":     true,
	"projects":          true,
	"languages":         true,
	"editors":           true,
	"categories":        true,
	"machines":          true,
	"operating_systems": true,
}

type insightsCollector struct {
	namedSeconds   *prometheus.Desc
	weekdaySeconds *prometheus.Desc
	daySeconds     *prometheus.Desc
	bestDay        *prometheus.Desc
	dailyAverage   *prometheus.Desc
	aiDaysRatio    *prometheus.Desc
	upToDate       *prometheus.Desc
	types          []string
	rangeName      string
	cacheTTL       time.Duration
	top            int
	minSeconds     float64
	uri            url.URL
	fetchStat      func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger         log.Logger

	mu      sync.Mutex
	cache   map[string]insightCacheEntry
	flights flightGroup
}

type insightCacheEntry struct {
	data  *wakatimeInsights
	until time.Time
}

func init() {
	registerCollector(insightsCollectorName, defaultDisabled, NewInsightsCollector)
}

func NewInsightsCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	types := in.Config.Insights.Types
	if len(types) == 0 {
		types = []string{"weekdays", "best_day", "daily_average", "ai_days"}
	}
	for _, insightType := range types {
		if !insightTypes[insightType] {
			return nil, fmt.Errorf("invalid insight type %q", insightType)
		}
	}

	rangeName := in.Config.Insights.Range
	if rangeName == "" {
		rangeName = defaultInsightsRange
	}
	if err := validateRange(rangeName); err != nil {
		return nil, fmt.Errorf("invalid insights range: %w", err)
	}

	cacheTTL := in.Config.Insights.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = time.Hour
	}

	return &insightsCollector{
		namedSeconds: newMetricDesc(
			"insight_seconds",
			"Coding seconds for a named insight dimension.",
			"range", "insight", "name",
		),
		weekdaySeconds: newMetricDesc(
			"insight_weekday_seconds",
			"Coding seconds per weekday.",
			"range", "weekday",
		),
		daySeconds: newMetricDesc(
			"insight_day_seconds",
			"Coding seconds per day for day-based insights.",
			"range", "insight", "date",
		),
		bestDay: newMetricDesc(
			"insight_best_day_seconds",
			"Coding seconds on the best day.",
			"range", "date",
		),
		dailyAverage: newMetricDesc(
			"insight_daily_average_seconds",
			"Average coding seconds per day.",
			"range",
		),
		aiDaysRatio: newMetricDesc(
			"insight_ai_days_ratio",
			"AI line-change percentage per day from the ai_days insight.",
			"range", "date",
		),
		upToDate: newMetricDesc(
			"insight_up_to_date",
			"Whether WakaTime reports the insight as up to date.",
			"range", "insight",
		),
		types:      types,
		rangeName:  rangeName,
		cacheTTL:   cacheTTL,
		top:        in.Config.Insights.Top,
		minSeconds: in.Config.Insights.MinSeconds,
		uri:        in.URI,
		fetchStat:  FetchHTTP(in.HTTPClient, in.Token, in.AuthMode, logger),
		logger:     logger,
		cache:      map[string]insightCacheEntry{},
	}, nil
}

func (c *insightsCollector) Update(ch chan<- prometheus.Metric) error {
	emitted := false
	var firstErr error

	for _, insightType := range c.types {
		data, err := c.getInsight(insightType)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			level.Error(c.logger).Log("msg", "Insight fetch failed", "insight", insightType, "err", err)
			continue
		}
		if c.emit(ch, insightType, data) {
			emitted = true
		}
	}

	if emitted {
		return nil
	}
	if firstErr != nil {
		return firstErr
	}
	return ErrNoData
}

func (c *insightsCollector) getInsight(insightType string) (*wakatimeInsights, error) {
	c.mu.Lock()
	if entry, ok := c.cache[insightType]; ok && time.Now().Before(entry.until) {
		data := entry.data
		c.mu.Unlock()
		return data, nil
	}
	c.mu.Unlock()

	return singleflightDo(&c.flights, insightType, func() (*wakatimeInsights, error) {
		body, fetchErr := c.fetchStat(c.uri, insightsEndpoint+"/"+insightType+"/"+c.rangeName, url.Values{})
		if fetchErr != nil {
			return nil, fetchErr
		}
		defer body.Close()

		var response wakatimeInsights
		if err := ReadAndUnmarshal(body, &response); err != nil {
			return nil, err
		}

		cacheTTL := c.cacheTTL
		if !response.IsUpToDate && !rawBool(response.Data, "is_up_to_date") {
			cacheTTL = time.Minute
		}

		c.mu.Lock()
		c.cache[insightType] = insightCacheEntry{data: &response, until: time.Now().Add(cacheTTL)}
		c.mu.Unlock()
		return &response, nil
	})
}

func (c *insightsCollector) emit(ch chan<- prometheus.Metric, insightType string, data *wakatimeInsights) bool {
	if insightType == "stats" {
		return false
	}

	raw, ok := data.Data[insightType]
	if !ok || len(raw) == 0 {
		return false
	}

	upToDate := rawBool(data.Data, "is_up_to_date") || data.IsUpToDate
	ch <- prometheus.MustNewConstMetric(c.upToDate, prometheus.GaugeValue, boolToFloat(upToDate), c.rangeName, insightType)

	switch insightType {
	case "weekdays":
		return c.emitWeekdays(ch, raw)
	case "best_day":
		return c.emitBestDay(ch, raw)
	case "daily_average":
		return c.emitDailyAverage(ch, raw)
	case "days", "ai_days":
		return c.emitDays(ch, insightType, raw)
	case "projects", "languages", "editors", "categories", "machines", "operating_systems":
		return c.emitNamed(ch, insightType, raw)
	default:
		return false
	}
}

func (c *insightsCollector) emitWeekdays(ch chan<- prometheus.Metric, raw json.RawMessage) bool {
	var named []insightNamed
	if json.Unmarshal(raw, &named) == nil && len(named) > 0 {
		for _, entry := range trimInsightNamed(named, c.top, c.minSeconds) {
			name := insightName(entry)
			if name == "" {
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.weekdaySeconds, prometheus.GaugeValue, entry.TotalSeconds, c.rangeName, name)
		}
		return true
	}

	var values map[string]float64
	if json.Unmarshal(raw, &values) == nil && len(values) > 0 {
		entries := make([]insightNamed, 0, len(values))
		for name, seconds := range values {
			entries = append(entries, insightNamed{Name: name, TotalSeconds: seconds})
		}
		for _, entry := range trimInsightNamed(entries, c.top, c.minSeconds) {
			ch <- prometheus.MustNewConstMetric(c.weekdaySeconds, prometheus.GaugeValue, entry.TotalSeconds, c.rangeName, entry.Name)
		}
		return true
	}
	return false
}

func (c *insightsCollector) emitBestDay(ch chan<- prometheus.Metric, raw json.RawMessage) bool {
	var best insightBestDay
	if json.Unmarshal(raw, &best) != nil || best.Date == "" {
		return false
	}
	ch <- prometheus.MustNewConstMetric(c.bestDay, prometheus.GaugeValue, best.TotalSeconds, c.rangeName, best.Date)
	return true
}

func (c *insightsCollector) emitDailyAverage(ch chan<- prometheus.Metric, raw json.RawMessage) bool {
	var entry insightBestDay
	if json.Unmarshal(raw, &entry) == nil {
		ch <- prometheus.MustNewConstMetric(c.dailyAverage, prometheus.GaugeValue, entry.TotalSeconds, c.rangeName)
		return true
	}

	var value float64
	if json.Unmarshal(raw, &value) == nil {
		ch <- prometheus.MustNewConstMetric(c.dailyAverage, prometheus.GaugeValue, value, c.rangeName)
		return true
	}
	return false
}

func (c *insightsCollector) emitDays(ch chan<- prometheus.Metric, insightType string, raw json.RawMessage) bool {
	var days []insightDay
	if json.Unmarshal(raw, &days) != nil || len(days) == 0 {
		return false
	}

	for _, day := range trimInsightDays(days, c.top, c.minSeconds) {
		if day.Date != "" {
			ch <- prometheus.MustNewConstMetric(c.daySeconds, prometheus.GaugeValue, day.TotalSeconds, c.rangeName, insightType, day.Date)
		}
		if insightType == "ai_days" {
			if ratio, ok := aiDaysRatio(day); ok {
				ch <- prometheus.MustNewConstMetric(c.aiDaysRatio, prometheus.GaugeValue, ratio, c.rangeName, day.Date)
			}
		}
	}
	return true
}

func (c *insightsCollector) emitNamed(ch chan<- prometheus.Metric, insightType string, raw json.RawMessage) bool {
	var named []insightNamed
	if json.Unmarshal(raw, &named) == nil && len(named) > 0 {
		for _, entry := range trimInsightNamed(named, c.top, c.minSeconds) {
			name := insightName(entry)
			if name == "" {
				continue
			}
			ch <- prometheus.MustNewConstMetric(c.namedSeconds, prometheus.GaugeValue, entry.TotalSeconds, c.rangeName, insightType, name)
		}
		return true
	}

	var values map[string]float64
	if json.Unmarshal(raw, &values) == nil && len(values) > 0 {
		entries := make([]insightNamed, 0, len(values))
		for name, seconds := range values {
			entries = append(entries, insightNamed{Name: name, TotalSeconds: seconds})
		}
		for _, entry := range trimInsightNamed(entries, c.top, c.minSeconds) {
			ch <- prometheus.MustNewConstMetric(c.namedSeconds, prometheus.GaugeValue, entry.TotalSeconds, c.rangeName, insightType, entry.Name)
		}
		return true
	}
	return false
}

func trimInsightNamed(entries []insightNamed, top int, minSeconds float64) []insightNamed {
	out := make([]insightNamed, 0, len(entries))
	for _, entry := range entries {
		if minSeconds > 0 && entry.TotalSeconds < minSeconds {
			continue
		}
		out = append(out, entry)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].TotalSeconds == out[j].TotalSeconds {
			return insightName(out[i]) < insightName(out[j])
		}
		return out[i].TotalSeconds > out[j].TotalSeconds
	})
	if top > 0 && len(out) > top {
		out = out[:top]
	}
	return out
}

func trimInsightDays(days []insightDay, top int, minSeconds float64) []insightDay {
	out := make([]insightDay, 0, len(days))
	for _, day := range days {
		if minSeconds > 0 && day.TotalSeconds < minSeconds {
			continue
		}
		out = append(out, day)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].TotalSeconds == out[j].TotalSeconds {
			return out[i].Date < out[j].Date
		}
		return out[i].TotalSeconds > out[j].TotalSeconds
	})
	if top > 0 && len(out) > top {
		out = out[:top]
	}
	return out
}

func insightName(entry insightNamed) string {
	if entry.Name != "" {
		return entry.Name
	}
	if entry.Day != "" {
		return entry.Day
	}
	return entry.Weekday
}

func aiDaysRatio(day insightDay) (float64, bool) {
	if day.AIPercent != 0 {
		return day.AIPercent, true
	}
	if day.AILinesPercent != 0 {
		return day.AILinesPercent, true
	}
	total := day.AILines + day.HumanLines
	if total > 0 {
		return float64(day.AILines) / float64(total) * 100, true
	}
	return 0, false
}

func rawBool(data map[string]json.RawMessage, key string) bool {
	raw, ok := data[key]
	if !ok {
		return false
	}
	var value bool
	if json.Unmarshal(raw, &value) != nil {
		return false
	}
	return value
}

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

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	aiCollectorName = "ai"
	aiSubsystem     = "ai"
	aiEndpoint      = "durations"
)

type aiCollector struct {
	lines             *prometheus.Desc
	tokens            *prometheus.Desc
	costUSD           *prometheus.Desc
	sessions          *prometheus.Desc
	promptEvents      *prometheus.Desc
	promptLengthSum   *prometheus.Desc
	promptLengthAvg   *prometheus.Desc
	secondsPerProject *prometheus.Desc
	uri               url.URL
	fetchStat         func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger            log.Logger
}

func init() {
	registerCollector(aiCollectorName, defaultEnabled, NewAICollector)
}

// NewAICollector returns a new Collector exposing AI coding stats from /durations.
// Requires OAuth scope read_heartbeats (broader than the other collectors).
func NewAICollector(in CommonInputs, logger log.Logger) (Collector, error) {
	return &aiCollector{
		lines: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, aiSubsystem, "lines_total"),
			"Total lines attributed to AI vs human typing, from Wakatime durations.",
			[]string{"origin", "kind"}, nil,
		),
		tokens: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, aiSubsystem, "tokens_total"),
			"Total AI tokens used, from Wakatime durations.",
			[]string{"kind"}, nil,
		),
		costUSD: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, aiSubsystem, "cost_usd_total"),
			"Estimated AI model cost in USD according to WakaTime pricing.",
			[]string{"model"}, nil,
		),
		sessions: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, aiSubsystem, "sessions_total"),
			"Sum of per-duration AI sessions, from Wakatime durations (may double-count overlapping sessions).",
			nil, nil,
		),
		promptEvents: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, aiSubsystem, "prompt_events_total"),
			"Total number of prompts sent to AI tools, from Wakatime durations.",
			nil, nil,
		),
		promptLengthSum: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, aiSubsystem, "prompt_length_chars_sum"),
			"Sum of user prompt characters typed to AI, from Wakatime durations.",
			nil, nil,
		),
		promptLengthAvg: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, aiSubsystem, "prompt_length_avg_chars"),
			"Mean prompt length in characters (sum / prompt events), from Wakatime durations.",
			nil, nil,
		),
		secondsPerProject: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, aiSubsystem, "seconds_total"),
			// Label is "name" for consistency with wakatime_project_seconds_total.
			"Total seconds per project, summed from Wakatime durations.",
			[]string{"name"}, nil,
		),
		uri:       in.URI,
		fetchStat: FetchHTTP(in.HTTPClient, in.Token, in.AuthMode, logger),
		logger:    logger,
	}, nil
}

func (c *aiCollector) Update(ch chan<- prometheus.Metric) error {
	params := url.Values{}
	params.Add("date", "today")
	params.Add("cache", "false")

	body, fetchErr := c.fetchStat(c.uri, aiEndpoint, params)
	if fetchErr != nil {
		return fetchErr
	}
	defer body.Close()

	durations := wakatimeDurations{}
	if err := ReadAndUnmarshal(body, &durations); err != nil {
		return err
	}

	if len(durations.Data) == 0 {
		return ErrNoData
	}

	level.Info(c.logger).Log(
		"msg", "Collecting AI stats from Wakatime",
		"durations", len(durations.Data),
		"start", durations.Start,
		"end", durations.End,
		"timezone", durations.Timezone,
	)

	var aiAdditions, aiDeletions, humanAdditions, humanDeletions int64
	var aiInputTokens, aiOutputTokens int64
	var aiSessions, aiPromptEvents int64
	var aiPromptLengthSum float64
	costsByModel := map[string]float64{}
	secondsByProject := map[string]float64{}

	for _, d := range durations.Data {
		aiAdditions += d.AIAdditions
		aiDeletions += d.AIDeletions
		humanAdditions += d.HumanAdditions
		humanDeletions += d.HumanDeletions
		aiInputTokens += d.AIInputTokens
		aiOutputTokens += d.AIOutputTokens
		aiSessions += d.AISessions
		aiPromptEvents += d.AIPromptEventsTotal
		aiPromptLengthSum += d.AIPromptLengthSum
		for model, cost := range d.AIModelCosts {
			costsByModel[model] += cost
		}
		if d.Project != "" {
			secondsByProject[d.Project] += d.Duration
		} else {
			secondsByProject["unknown"] += d.Duration
		}
	}

	ch <- prometheus.MustNewConstMetric(c.lines, prometheus.CounterValue, float64(aiAdditions), "ai", "additions")
	ch <- prometheus.MustNewConstMetric(c.lines, prometheus.CounterValue, float64(aiDeletions), "ai", "deletions")
	ch <- prometheus.MustNewConstMetric(c.lines, prometheus.CounterValue, float64(humanAdditions), "human", "additions")
	ch <- prometheus.MustNewConstMetric(c.lines, prometheus.CounterValue, float64(humanDeletions), "human", "deletions")

	ch <- prometheus.MustNewConstMetric(c.tokens, prometheus.CounterValue, float64(aiInputTokens), "input")
	ch <- prometheus.MustNewConstMetric(c.tokens, prometheus.CounterValue, float64(aiOutputTokens), "output")

	for model, cost := range costsByModel {
		ch <- prometheus.MustNewConstMetric(c.costUSD, prometheus.CounterValue, cost, model)
	}

	ch <- prometheus.MustNewConstMetric(c.sessions, prometheus.CounterValue, float64(aiSessions))
	ch <- prometheus.MustNewConstMetric(c.promptEvents, prometheus.CounterValue, float64(aiPromptEvents))
	ch <- prometheus.MustNewConstMetric(c.promptLengthSum, prometheus.CounterValue, float64(aiPromptLengthSum))

	var promptLengthAvg float64
	if aiPromptEvents > 0 {
		promptLengthAvg = aiPromptLengthSum / float64(aiPromptEvents)
	}
	ch <- prometheus.MustNewConstMetric(c.promptLengthAvg, prometheus.GaugeValue, promptLengthAvg)

	for project, seconds := range secondsByProject {
		ch <- prometheus.MustNewConstMetric(c.secondsPerProject, prometheus.CounterValue, seconds, project)
	}

	return nil
}

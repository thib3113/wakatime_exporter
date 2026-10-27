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
	"context"
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
	top               int
	minSeconds        float64
	uri               url.URL
	fetchStat         func(context.Context, url.URL, string, url.Values) (io.ReadCloser, error)
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
		secondsPerProject: newMetricDesc(
			"durations_seconds_total",
			"Total seconds per project, summed from Wakatime durations.",
			"name",
		),
		top:        in.Config.AI.Top,
		minSeconds: in.Config.AI.MinSeconds,
		uri:        in.URI,
		fetchStat:  FetchHTTP(in.HTTPClient, in.Token, in.AuthMode, logger),
		logger:     logger,
	}, nil
}

func (c *aiCollector) UpdateContext(ctx context.Context, ch chan<- prometheus.Metric) error {
	params := url.Values{}
	params.Add("date", "today")
	params.Add("cache", "false")

	body, fetchErr := c.fetchStat(ctx, c.uri, aiEndpoint, params)
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

	totals := newAITotals()
	secondsByProject := map[string]float64{}

	for _, d := range durations.Data {
		totals.addDuration(d)
		if d.Project != "" {
			secondsByProject[d.Project] += d.Duration
		} else {
			secondsByProject["unknown"] += d.Duration
		}
	}

	ch <- prometheus.MustNewConstMetric(c.lines, prometheus.CounterValue, float64(totals.additions), "ai", "additions")
	ch <- prometheus.MustNewConstMetric(c.lines, prometheus.CounterValue, float64(totals.deletions), "ai", "deletions")
	ch <- prometheus.MustNewConstMetric(c.lines, prometheus.CounterValue, float64(totals.humanAdditions), "human", "additions")
	ch <- prometheus.MustNewConstMetric(c.lines, prometheus.CounterValue, float64(totals.humanDeletions), "human", "deletions")

	ch <- prometheus.MustNewConstMetric(c.tokens, prometheus.CounterValue, float64(totals.inputTokens), "input")
	ch <- prometheus.MustNewConstMetric(c.tokens, prometheus.CounterValue, float64(totals.outputTokens), "output")

	models := topNamedValues(totals.costs, c.top, 0)
	warnDropped(c.logger, "ai", "model", len(totals.costs), len(models))
	for _, model := range models {
		ch <- prometheus.MustNewConstMetric(c.costUSD, prometheus.CounterValue, model.value, model.name)
	}

	ch <- prometheus.MustNewConstMetric(c.sessions, prometheus.CounterValue, float64(totals.sessions))
	ch <- prometheus.MustNewConstMetric(c.promptEvents, prometheus.CounterValue, float64(totals.promptEvents))
	ch <- prometheus.MustNewConstMetric(c.promptLengthSum, prometheus.CounterValue, totals.promptLengthSum)

	var promptLengthAvg float64
	if totals.promptEvents > 0 {
		promptLengthAvg = totals.promptLengthSum / float64(totals.promptEvents)
	}
	ch <- prometheus.MustNewConstMetric(c.promptLengthAvg, prometheus.GaugeValue, promptLengthAvg)

	projects := topNamedValues(secondsByProject, c.top, c.minSeconds)
	warnDropped(c.logger, "ai", "project", len(secondsByProject), len(projects))
	for _, project := range projects {
		ch <- prometheus.MustNewConstMetric(c.secondsPerProject, prometheus.CounterValue, project.value, project.name)
	}

	return nil
}

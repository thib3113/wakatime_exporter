/*
Copyright 2020 Jacob Colvin (MacroPower)
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

package main

import (
	"crypto/tls"
	"html"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/common/version"

	"github.com/thib3113/wakatime_exporter/collector"
)

// newLogger builds a go-kit logger with the given level and format.
// It replaces promlog, which was removed from prometheus/common.
func newLogger(lvl, format string) log.Logger {
	w := log.NewSyncWriter(os.Stderr)
	var logger log.Logger
	if format == "json" {
		logger = log.NewJSONLogger(w)
	} else {
		logger = log.NewLogfmtLogger(w)
	}
	var opt level.Option
	switch strings.ToLower(lvl) {
	case "debug":
		opt = level.AllowDebug()
	case "warn":
		opt = level.AllowWarn()
	case "error":
		opt = level.AllowError()
	default:
		opt = level.AllowInfo()
	}
	return log.With(level.NewFilter(logger, opt), "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
}

// UserPath appends the User path to a given URL
func UserPath(uri *url.URL, user string) url.URL {
	userURL := *uri
	userPath := path.Join(userURL.Path, "users", user)

	userURL.Path = userPath
	return userURL
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		out = append(out, part)
	}
	return out
}

func main() {
	var (
		disableDefaultCollectors = kingpin.Flag(
			"collector.disable-defaults",
			"Set all collectors to disabled by default.",
		).Default("false").Envar("WAKA_DISABLE_DEFAULT_COLLECTORS").Bool()

		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address to listen on for web interface and metrics.",
		).Default(":9212").Envar("WAKA_LISTEN_ADDRESS").String()

		metricsPath = kingpin.Flag(
			"web.metrics-path",
			"Path under which to expose metrics.",
		).Default("/metrics").Envar("WAKA_METRICS_PATH").String()

		disableExporterMetrics = kingpin.Flag(
			"web.disable-exporter-metrics",
			"Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).",
		).Default("false").Envar("WAKA_DISABLE_EXPORTER_METRICS").Bool()

		wakaScrapeURI = kingpin.Flag(
			"wakatime.scrape-uri",
			"Base path to query for Wakatime data.",
		).Default("https://wakatime.com/api/v1").Envar("WAKA_SCRAPE_URI").String()

		wakaUser = kingpin.Flag(
			"wakatime.user",
			"User to query for Wakatime data.",
		).Default("current").Envar("WAKA_USER").String()

		wakaToken = kingpin.Flag(
			"wakatime.api-key",
			"Token to use when getting stats from Wakatime. Accepts a classic API key (Basic auth) or a WakaTime OAuth access token (Bearer auth, waka_tok_...).",
		).Required().Envar("WAKA_API_KEY").String()

		wakaAuthMode = kingpin.Flag(
			"wakatime.auth-mode",
			"Auth mode for the Wakatime API: auto (Bearer for waka_tok_ tokens, Basic otherwise), basic, or bearer.",
		).Default("auto").Envar("WAKA_AUTH_MODE").Enum("auto", "basic", "bearer")

		wakaTimeout = kingpin.Flag(
			"wakatime.timeout",
			"Timeout for trying to get stats from Wakatime.",
		).Default("5s").Envar("WAKA_TIMEOUT").Duration()

		wakaSSLVerify = kingpin.Flag(
			"wakatime.ssl-verify",
			"Flag that enables SSL certificate verification for the scrape URI.",
		).Default("true").Envar("WAKA_SSL_VERIFY").Bool()

		statsRange = kingpin.Flag(
			"stats.range",
			"Time range for the stats collector (all_time, last_7_days, last_30_days, last_6_months, last_year, YYYY, or YYYY-MM).",
		).Default("all_time").Envar("WAKA_STATS_RANGE").String()

		statsCacheTTL = kingpin.Flag(
			"stats.cache-ttl",
			"How long to cache the stats response between scrapes.",
		).Default("1h").Envar("WAKA_STATS_CACHE_TTL").Duration()

		statsTop = kingpin.Flag(
			"stats.top",
			"Keep at most N entries per stats dimension (0 = all).",
		).Default("50").Envar("WAKA_STATS_TOP").Int()

		statsMinSeconds = kingpin.Flag(
			"stats.min-seconds",
			"Drop stats entries below this many seconds (0 = keep all).",
		).Default("60").Envar("WAKA_STATS_MIN_SECONDS").Float64()

		aiBreakdown = kingpin.Flag(
			"ai.breakdown",
			"Comma-separated durations slice_by dimensions for AI breakdown metrics (language, editor, os, category, machine, project, entity, dependencies).",
		).Default("language").Envar("WAKA_AI_BREAKDOWN").String()

		aiTop = kingpin.Flag(
			"ai.top",
			"Keep at most N entries per AI breakdown dimension (0 = all).",
		).Default("50").Envar("WAKA_AI_TOP").Int()

		aiMinTokens = kingpin.Flag(
			"ai.min-tokens",
			"Drop AI breakdown entries below this many total tokens (0 = keep all).",
		).Default("1000").Envar("WAKA_AI_MIN_TOKENS").Int64()

		aiCacheTTL = kingpin.Flag(
			"ai.cache-ttl",
			"How long to cache AI breakdown responses between scrapes.",
		).Default("15m").Envar("WAKA_AI_CACHE_TTL").Duration()

		insightsTypes = kingpin.Flag(
			"insights.types",
			"Comma-separated insight types (weekdays, days, ai_days, best_day, daily_average, projects, languages, editors, categories, machines, operating_systems, stats).",
		).Default("weekdays,best_day,daily_average,ai_days").Envar("WAKA_INSIGHTS_TYPES").String()

		insightsRange = kingpin.Flag(
			"insights.range",
			"Time range for insights (all_time, last_7_days, last_30_days, last_6_months, last_year, YYYY, or YYYY-MM).",
		).Default("all_time").Envar("WAKA_INSIGHTS_RANGE").String()

		insightsCacheTTL = kingpin.Flag(
			"insights.cache-ttl",
			"How long to cache insights responses between scrapes.",
		).Default("1h").Envar("WAKA_INSIGHTS_CACHE_TTL").Duration()

		insightsTop = kingpin.Flag(
			"insights.top",
			"Keep at most N entries per insights dimension (0 = all).",
		).Default("50").Envar("WAKA_INSIGHTS_TOP").Int()

		insightsMinSeconds = kingpin.Flag(
			"insights.min-seconds",
			"Drop insights entries below this many seconds (0 = keep all).",
		).Default("60").Envar("WAKA_INSIGHTS_MIN_SECONDS").Float64()

		inventoryCacheTTL = kingpin.Flag(
			"inventory.cache-ttl",
			"How long to cache user agents, projects and machines between scrapes.",
		).Default("1h").Envar("WAKA_INVENTORY_CACHE_TTL").Duration()

		projectsFilter = kingpin.Flag(
			"projects.filter",
			"Only export projects whose name matches this regex (empty = all).",
		).Default("").Envar("WAKA_PROJECTS_FILTER").String()

		statsInclude = kingpin.Flag(
			"stats.include",
			"Regex matched against dimension:name for stats (project:, language:, editor:, category:, os:, machine:, dependency:).",
		).Default("").Envar("WAKA_STATS_INCLUDE").String()

		statsExclude = kingpin.Flag(
			"stats.exclude",
			"Regex matched against dimension:name for stats; matching entries are dropped.",
		).Default("").Envar("WAKA_STATS_EXCLUDE").String()

		inventoryTop = kingpin.Flag(
			"inventory.top",
			"Keep at most N user agents, projects and machines by recency (0 = all).",
		).Default("0").Envar("WAKA_INVENTORY_TOP").Int()

		inventoryInclude = kingpin.Flag(
			"inventory.include",
			"Regex matched against editor:<editor>, project:<name>, machine:<name> for inventory collectors.",
		).Default("").Envar("WAKA_INVENTORY_INCLUDE").String()

		inventoryExclude = kingpin.Flag(
			"inventory.exclude",
			"Regex matched against editor:<editor>, project:<name>, machine:<name>; matching entries are dropped.",
		).Default("").Envar("WAKA_INVENTORY_EXCLUDE").String()

		aiInclude = kingpin.Flag(
			"ai.include",
			"Regex matched against <slice>:<name> for AI breakdown metrics.",
		).Default("").Envar("WAKA_AI_INCLUDE").String()

		aiExclude = kingpin.Flag(
			"ai.exclude",
			"Regex matched against <slice>:<name>; matching AI breakdown entries are dropped.",
		).Default("").Envar("WAKA_AI_EXCLUDE").String()

		heartbeatsInclude = kingpin.Flag(
			"heartbeats.include",
			"Regex matched against project:<name> for heartbeat metrics.",
		).Default("").Envar("WAKA_HEARTBEATS_INCLUDE").String()

		heartbeatsExclude = kingpin.Flag(
			"heartbeats.exclude",
			"Regex matched against project:<name>; matching heartbeat projects are dropped.",
		).Default("").Envar("WAKA_HEARTBEATS_EXCLUDE").String()

		externalDurationsInclude = kingpin.Flag(
			"external-durations.include",
			"Regex matched against project:, provider:, category:, language: for external durations.",
		).Default("").Envar("WAKA_EXTERNAL_DURATIONS_INCLUDE").String()

		externalDurationsExclude = kingpin.Flag(
			"external-durations.exclude",
			"Regex matched against project:, provider:, category:, language:; matching entries are dropped.",
		).Default("").Envar("WAKA_EXTERNAL_DURATIONS_EXCLUDE").String()

		externalDurationsDate = kingpin.Flag(
			"external-durations.date",
			"Date for the external durations collector (today, yesterday, or YYYY-MM-DD).",
		).Default("today").Envar("WAKA_EXTERNAL_DURATIONS_DATE").String()

		externalDurationsCacheTTL = kingpin.Flag(
			"external-durations.cache-ttl",
			"How long to cache external durations between scrapes.",
		).Default("15m").Envar("WAKA_EXTERNAL_DURATIONS_CACHE_TTL").Duration()

		privateLeaderboardsBoards = kingpin.Flag(
			"private-leaderboards.boards",
			"Comma-separated private leaderboard IDs (empty = all).",
		).Default("").Envar("WAKA_PRIVATE_LEADERBOARDS_BOARDS").String()

		privateLeaderboardsTypes = kingpin.Flag(
			"private-leaderboards.types",
			"Comma-separated leaderboard types (time, manual, ai, spend).",
		).Default("time,ai,spend").Envar("WAKA_PRIVATE_LEADERBOARDS_TYPES").String()

		privateLeaderboardsMaxPages = kingpin.Flag(
			"private-leaderboards.max-pages",
			"Maximum number of pages to fetch per leaderboard (1 = first page only).",
		).Default("1").Envar("WAKA_PRIVATE_LEADERBOARDS_MAX_PAGES").Int()

		privateLeaderboardsCacheTTL = kingpin.Flag(
			"private-leaderboards.cache-ttl",
			"How long to cache private leaderboards between scrapes.",
		).Default("1h").Envar("WAKA_PRIVATE_LEADERBOARDS_CACHE_TTL").Duration()

		orgsIDs = kingpin.Flag(
			"orgs.ids",
			"Comma-separated organization IDs (empty = all).",
		).Default("").Envar("WAKA_ORGS_IDS").String()

		orgsCacheTTL = kingpin.Flag(
			"orgs.cache-ttl",
			"How long to cache organizations between scrapes.",
		).Default("1h").Envar("WAKA_ORGS_CACHE_TTL").Duration()

		heartbeatsTop = kingpin.Flag(
			"heartbeats.top",
			"Keep at most N heartbeats projects (0 = all).",
		).Default("50").Envar("WAKA_HEARTBEATS_TOP").Int()

		heartbeatsMinCount = kingpin.Flag(
			"heartbeats.min-count",
			"Drop heartbeat projects below this many heartbeats (0 = keep all).",
		).Default("1").Envar("WAKA_HEARTBEATS_MIN_COUNT").Int()

		heartbeatsCacheTTL = kingpin.Flag(
			"heartbeats.cache-ttl",
			"How long to cache heartbeats between scrapes.",
		).Default("15m").Envar("WAKA_HEARTBEATS_CACHE_TTL").Duration()
		logLevel = kingpin.Flag(
			"log.level",
			"Only log messages with the given severity or above. One of: [debug, info, warn, error]",
		).Default("info").Envar("WAKA_LOG_LEVEL").Enum("debug", "info", "warn", "error")

		logFormat = kingpin.Flag(
			"log.format",
			"Output format of log messages. One of: [logfmt, json]",
		).Default("logfmt").Envar("WAKA_LOG_FORMAT").Enum("logfmt", "json")
	)

	kingpin.Version(version.Print("wakatime_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := newLogger(*logLevel, *logFormat)

	if *disableDefaultCollectors {
		collector.DisableDefaultCollectors()
	}
	level.Info(logger).Log("msg", "Starting wakatime_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())
	level.Info(logger).Log("msg", "Wakatime auth mode", "mode", collector.EffectiveAuthMode(*wakaToken, *wakaAuthMode))

	wakaBaseURI, err := url.Parse(*wakaScrapeURI)
	if err != nil {
		level.Error(logger).Log("msg", "Error parsing URL", "err", err)
		os.Exit(1)
	}

	if !strings.HasPrefix(*metricsPath, "/") || strings.ContainsAny(*metricsPath, "{} \t\r\n") {
		level.Error(logger).Log("msg", "Invalid metrics path", "path", *metricsPath)
		os.Exit(1)
	}

	httpClient := &http.Client{
		Timeout: *wakaTimeout,
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: !*wakaSSLVerify},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 20,
			IdleConnTimeout:     90 * time.Second,
			ForceAttemptHTTP2:   true,
		},
	}

	mux := http.NewServeMux()
	mux.Handle(*metricsPath, newHandler(collector.CommonInputs{
		BaseURI:    *wakaBaseURI,
		URI:        UserPath(wakaBaseURI, *wakaUser),
		Token:      *wakaToken,
		AuthMode:   *wakaAuthMode,
		SSLVerify:  *wakaSSLVerify,
		Timeout:    *wakaTimeout,
		HTTPClient: httpClient,
		Config: collector.Config{
			Stats: collector.StatsConfig{
				Range:      *statsRange,
				CacheTTL:   *statsCacheTTL,
				Top:        *statsTop,
				MinSeconds: *statsMinSeconds,
				Filter:     collector.FilterConfig{Include: *statsInclude, Exclude: *statsExclude},
			},
			AI: collector.AIBreakdownConfig{
				Dimensions: splitCSV(*aiBreakdown),
				Top:        *aiTop,
				MinTokens:  *aiMinTokens,
				CacheTTL:   *aiCacheTTL,
				Filter:     collector.FilterConfig{Include: *aiInclude, Exclude: *aiExclude},
			},
			Insights: collector.InsightsConfig{
				Types:      splitCSV(*insightsTypes),
				Range:      *insightsRange,
				CacheTTL:   *insightsCacheTTL,
				Top:        *insightsTop,
				MinSeconds: *insightsMinSeconds,
			},
			Inventory: collector.InventoryConfig{
				Top:            *inventoryTop,
				Filter:         collector.FilterConfig{Include: *inventoryInclude, Exclude: *inventoryExclude},
				ProjectsFilter: *projectsFilter,
				CacheTTL:       *inventoryCacheTTL,
			},
			ExternalDurations: collector.ExternalDurationsConfig{
				Date:     *externalDurationsDate,
				CacheTTL: *externalDurationsCacheTTL,
				Filter:   collector.FilterConfig{Include: *externalDurationsInclude, Exclude: *externalDurationsExclude},
			},
			PrivateLeaderboards: collector.PrivateLeaderboardsConfig{
				Boards:   splitCSV(*privateLeaderboardsBoards),
				Types:    splitCSV(*privateLeaderboardsTypes),
				MaxPages: *privateLeaderboardsMaxPages,
				CacheTTL: *privateLeaderboardsCacheTTL,
			},
			Orgs: collector.OrgsConfig{
				IDs:      splitCSV(*orgsIDs),
				CacheTTL: *orgsCacheTTL,
			},
			Heartbeats: collector.HeartbeatsConfig{
				Top:      *heartbeatsTop,
				MinCount: *heartbeatsMinCount,
				CacheTTL: *heartbeatsCacheTTL,
				Filter:   collector.FilterConfig{Include: *heartbeatsInclude, Exclude: *heartbeatsExclude},
			},
		},
	}, !*disableExporterMetrics, logger))
	if *metricsPath != "/" {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html>
				<head><title>Wakatime Exporter</title></head>
				<body>
				<h1>Wakatime Exporter</h1>
				<p><a href="` + html.EscapeString(*metricsPath) + `">Metrics</a></p>
				</body>
				</html>`))
		})
	}

	level.Info(logger).Log("msg", "Listening on", "address", *listenAddress)
	server := &http.Server{
		Addr:              *listenAddress,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}

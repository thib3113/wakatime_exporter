<a href="#"><img src="https://raw.githubusercontent.com/thib3113/wakatime_exporter/main/docs/img/waka_header.png"></a>

<h2></h2>
<p align="center">
  <a href="https://github.com/thib3113/wakatime_exporter/blob/main/go.mod"><img alt="GitHub go.mod Go version" src="https://img.shields.io/github/go-mod/go-version/thib3113/wakatime_exporter"></a>
  <a href="https://hub.docker.com/r/thib3113/wakatime-exporter"><img alt="Docker Image Size (latest by date)" src="https://img.shields.io/docker/image-size/thib3113/wakatime-exporter?color=green"></a>
  <a href="https://hub.docker.com/r/thib3113/wakatime-exporter"><img alt="Docker Pulls" src="https://img.shields.io/docker/pulls/thib3113/wakatime-exporter"></a>
</p>

_wakatime_exporter_ is a Prometheus exporter for [Wakatime](https://wakatime.com/) statistics.
It intends to extend the existing Wakatime ecosystem
via allowing users to make use of Prometheus (and therefore any consumers of its API)
as companion services alongside the traditional Wakatime web application.
This could be anything from including some of your coding statistics in Grafana,
to forecasting your coding time with prophet.

> **Actively maintained**: this repository ([thib3113/wakatime_exporter](https://github.com/thib3113/wakatime_exporter))
> continues the project [MacroPower/wakatime_exporter](https://github.com/MacroPower/wakatime_exporter),
> which appears unmaintained (last upstream commit: February 2023).
> It is published as a standalone repository (not a GitHub fork) so it can be found via GitHub search.
> Many thanks to Jacob Colvin (MacroPower) for creating the project and maintaining it up to v0.1.0 —
> all of his work is preserved as-is, including the full git history and the original license headers.
> On top of upstream, this repository adds:
> OAuth 2.0 Bearer authentication (required since WakaTime API keys become write-only on 2026-11-01),
> a new `ai` collector exposing AI coding statistics from the `/durations` endpoint,
> a `stats` collector exposing range totals (including all-time per-project totals)
> from the `/stats/{range}` endpoint, and additional collectors for AI breakdowns,
> insights, external durations, user agents, projects, machines, private leaderboards,
> organizations and heartbeats.

> NOTE: _wakatime_exporter_ is currently in ALPHA. Expect things to break and change.

[Click here](docs/METRICS.md) to see an example of the exported metrics.

[Click here](https://grafana.com/grafana/dashboards/12790) for a simple dashboard you can use to get started.

<a href="#"><img src="https://raw.githubusercontent.com/thib3113/wakatime_exporter/main/docs/img/example.png"></a>

## Usage

In most cases, you should only need to provide an API key.
All other parameters are for advanced use-cases only
and you should be able to leave them set to their defaults.

You can get your Wakatime API key by visiting: https://wakatime.com/api-key

Provide arguments via parameters:

```text
usage: wakatime_exporter --wakatime.api-key=WAKATIME.API-KEY [<flags>]

Flags:
  --help, -h                     Show context-sensitive help.
  --collector.ai                     Enable the ai collector (default: enabled).
  --collector.ai-breakdown           Enable the ai-breakdown collector (default: enabled).
  --collector.all-time               Enable the all-time collector (default: enabled).
  --collector.external-durations     Enable the external-durations collector (default: enabled).
  --collector.goal                   Enable the goal collector (default: enabled).
  --collector.heartbeats             Enable the heartbeats collector (default: disabled).
  --collector.insights               Enable the insights collector (default: disabled).
  --collector.leader                 Enable the leader collector (default: enabled).
  --collector.machines               Enable the machines collector (default: enabled).
  --collector.orgs                   Enable the orgs collector (default: disabled).
  --collector.private-leaderboards   Enable the private-leaderboards collector (default: disabled).
  --collector.projects               Enable the projects collector (default: enabled).
  --collector.stats                  Enable the stats collector (default: enabled).
  --collector.summary                Enable the summary collector (default: enabled).
  --collector.user-agents            Enable the user-agents collector (default: enabled).
  --collector.disable-defaults       Set all collectors to disabled by default.
  --web.listen-address=":9212"   Address to listen on for web interface and telemetry.
  --web.metrics-path="/metrics"  Path under which to expose metrics.
  --web.disable-exporter-metrics Exclude metrics about the exporter itself (promhttp_*, process_*, go_*).
  --wakatime.scrape-uri="https://wakatime.com/api/v1"
                                 Base path to query for Wakatime data.
  --wakatime.user="current"      User to query for Wakatime data.
  --wakatime.api-key             Token to use when getting stats from Wakatime. Accepts a classic API key (Basic auth)
                                 or a WakaTime OAuth access token (Bearer auth, waka_tok_...).
  --wakatime.auth-mode="auto"    Auth mode for the Wakatime API: auto (Bearer for waka_tok_ tokens, Basic otherwise),
                                 basic, or bearer.
  --wakatime.timeout=5s          Timeout for trying to get stats from Wakatime.
  --wakatime.ssl-verify          Flag that enables SSL certificate verification for the scrape URI.
  --stats.range="all_time"           Time range for the stats collector: all_time, last_7_days, last_30_days,
                                     last_6_months, last_year, YYYY, or YYYY-MM.
  --stats.cache-ttl=1h               How long to cache the stats response between scrapes.
  --stats.top=50                     Keep at most N entries per stats dimension (0 = all).
  --stats.min-seconds=60             Drop stats entries below this many seconds (0 = keep all).
  --stats.include=REGEX              Regex matched against dimension:name for stats.
  --stats.exclude=REGEX              Regex matched against dimension:name; matching entries are dropped.
  --ai.breakdown="language"          Comma-separated durations slice_by dimensions for AI breakdown
                                     (language, editor, os, category, machine, project, entity, dependencies).
  --ai.top=50                        Keep at most N entries per AI breakdown dimension (0 = all).
  --ai.min-tokens=1000               Drop AI breakdown entries below this many total tokens (0 = keep all).
  --ai.cache-ttl=15m                 How long to cache AI breakdown responses between scrapes.
  --ai.include=REGEX                 Regex matched against <slice>:<name> for AI breakdown metrics.
  --ai.exclude=REGEX                 Regex matched against <slice>:<name>; matching entries are dropped.
  --insights.types="weekdays,best_day,daily_average,ai_days"
                                     Comma-separated insight types.
  --insights.range="all_time"        Time range for insights.
  --insights.cache-ttl=1h            How long to cache insights responses between scrapes.
  --insights.top=50                  Keep at most N entries per insights dimension (0 = all).
  --insights.min-seconds=60          Drop insights entries below this many seconds (0 = keep all).
  --inventory.top=0                  Keep at most N user agents/projects/machines by recency (0 = all).
  --inventory.include=REGEX          Regex matched against editor:, project:, machine: for inventory.
  --inventory.exclude=REGEX          Regex matched against editor:, project:, machine:; matching entries are dropped.
  --inventory.cache-ttl=1h           Cache duration for user agents, projects and machines.
  --projects.filter=""               Only export projects whose name matches this regex.
  --external-durations.date="today"  Date for the external durations collector.
  --external-durations.include=REGEX Regex matched against project:, provider:, category:, language:.
  --external-durations.exclude=REGEX Regex matched against project:, provider:, category:, language:; matching entries are dropped.
  --external-durations.cache-ttl=15m How long to cache external durations between scrapes.
  --private-leaderboards.boards=""   Comma-separated private leaderboard IDs (empty = all).
  --private-leaderboards.types="time,ai,spend"
                                     Comma-separated leaderboard types.
  --private-leaderboards.max-pages=1 Maximum number of pages per leaderboard.
  --private-leaderboards.cache-ttl=1h
                                     How long to cache private leaderboards between scrapes.
  --orgs.ids=""                      Comma-separated organization IDs (empty = all).
  --orgs.cache-ttl=1h                How long to cache organizations between scrapes.
  --heartbeats.top=50                Keep at most N heartbeat projects (0 = all).
  --heartbeats.min-count=1           Drop heartbeat projects below this many heartbeats (0 = keep all).
  --heartbeats.include=REGEX         Regex matched against project:<name> for heartbeat metrics.
  --heartbeats.exclude=REGEX         Regex matched against project:<name>; matching projects are dropped.
  --heartbeats.cache-ttl=15m         How long to cache heartbeats between scrapes.
  --log.level=info               Only log messages with the given severity or above.
                                 One of: [debug, info, warn, error]
  --log.format=logfmt            Output format of log messages.
                                 One of: [logfmt, json]
  --version                      Show application version.
```

and/or via environment variables:

```shell
WAKA_LISTEN_ADDRESS=":9212"                   # Address to listen on for web interface and telemetry.
WAKA_METRICS_PATH="/metrics"                  # Path under which to expose metrics.
WAKA_SCRAPE_URI="https://wakatime.com/api/v1" # Base path to query for Wakatime data.
WAKA_USER="current"                           # User to query for Wakatime data.
WAKA_API_KEY=""                               # Token to use when getting stats from Wakatime.
WAKA_AUTH_MODE="auto"                         # Auth mode: auto, basic, or bearer.
WAKA_TIMEOUT="5s"                             # Timeout for trying to get stats from Wakatime.
WAKA_SSL_VERIFY="true"                        # SSL certificate verification for the scrape URI.
WAKA_DISABLE_EXPORTER_METRICS="false"         # Exclude metrics about the exporter itself.
WAKA_DISABLE_DEFAULT_COLLECTORS="false"       # Start with all collectors disabled.
WAKA_COLLECTOR_AI="true"                      # Enable the ai collector.
WAKA_COLLECTOR_AIBREAKDOWN="true"             # Enable the ai-breakdown collector.
WAKA_COLLECTOR_ALLTIME="true"                 # Enable the all-time collector.
WAKA_COLLECTOR_EXTERNALDURATIONS="true"       # Enable the external-durations collector.
WAKA_COLLECTOR_GOAL="true"                    # Enable the goal collector.
WAKA_COLLECTOR_HEARTBEATS="false"             # Enable the heartbeats collector.
WAKA_COLLECTOR_INSIGHTS="false"               # Enable the insights collector.
WAKA_COLLECTOR_LEADER="true"                  # Enable the leader collector.
WAKA_COLLECTOR_MACHINES="true"                # Enable the machines collector.
WAKA_COLLECTOR_ORGS="false"                   # Enable the orgs collector.
WAKA_COLLECTOR_PRIVATELEADERBOARDS="false"    # Enable the private-leaderboards collector.
WAKA_COLLECTOR_PROJECTS="true"                # Enable the projects collector.
WAKA_COLLECTOR_STATS="true"                   # Enable the stats collector.
WAKA_COLLECTOR_SUMMARY="true"                 # Enable the summary collector.
WAKA_COLLECTOR_USERAGENTS="true"              # Enable the user-agents collector.
WAKA_STATS_RANGE="all_time"                   # Time range for the stats collector.
WAKA_STATS_CACHE_TTL="1h"                     # Cache duration for the stats response.
WAKA_STATS_TOP="50"                           # Max entries per stats dimension (0 = all).
WAKA_STATS_MIN_SECONDS="60"                   # Drop stats entries below this many seconds.
WAKA_STATS_INCLUDE=""                         # Regex on dimension:name for stats (include).
WAKA_STATS_EXCLUDE=""                         # Regex on dimension:name for stats (exclude).
WAKA_AI_BREAKDOWN="language"                  # AI breakdown dimensions (comma-separated).
WAKA_AI_TOP="50"                              # Max entries per AI breakdown dimension (0 = all).
WAKA_AI_MIN_TOKENS="1000"                     # Drop AI breakdown entries below this many tokens.
WAKA_AI_CACHE_TTL="15m"                       # Cache duration for AI breakdown responses.
WAKA_AI_INCLUDE=""                            # Regex on <slice>:<name> for AI breakdown (include).
WAKA_AI_EXCLUDE=""                            # Regex on <slice>:<name> for AI breakdown (exclude).
WAKA_INSIGHTS_TYPES="weekdays,best_day,daily_average,ai_days"
WAKA_INSIGHTS_RANGE="all_time"                # Time range for insights.
WAKA_INSIGHTS_CACHE_TTL="1h"                  # Cache duration for insights responses.
WAKA_INSIGHTS_TOP="50"                        # Max entries per insights dimension.
WAKA_INSIGHTS_MIN_SECONDS="60"                # Drop insights entries below this many seconds.
WAKA_INVENTORY_TOP="0"                        # Keep N most recent user agents/projects/machines.
WAKA_INVENTORY_INCLUDE=""                     # Regex on editor:, project:, machine: (include).
WAKA_INVENTORY_EXCLUDE=""                     # Regex on editor:, project:, machine: (exclude).
WAKA_INVENTORY_CACHE_TTL="1h"                 # Cache duration for user agents, projects, machines.
WAKA_PROJECTS_FILTER=""                       # Only export projects matching this regex.
WAKA_EXTERNAL_DURATIONS_DATE="today"          # Date for external durations.
WAKA_EXTERNAL_DURATIONS_INCLUDE=""            # Regex on project:, provider:, category:, language: (include).
WAKA_EXTERNAL_DURATIONS_EXCLUDE=""            # Regex on project:, provider:, category:, language: (exclude).
WAKA_EXTERNAL_DURATIONS_CACHE_TTL="15m"       # Cache duration for external durations.
WAKA_PRIVATE_LEADERBOARDS_BOARDS=""           # Leaderboard IDs (empty = all).
WAKA_PRIVATE_LEADERBOARDS_TYPES="time,ai,spend"
WAKA_PRIVATE_LEADERBOARDS_MAX_PAGES="1"       # Max pages per leaderboard.
WAKA_PRIVATE_LEADERBOARDS_CACHE_TTL="1h"      # Cache duration for private leaderboards.
WAKA_ORGS_IDS=""                              # Organization IDs (empty = all).
WAKA_ORGS_CACHE_TTL="1h"                      # Cache duration for organizations.
WAKA_HEARTBEATS_TOP="50"                      # Max heartbeat projects (0 = all).
WAKA_HEARTBEATS_MIN_COUNT="1"                 # Drop projects below this many heartbeats.
WAKA_HEARTBEATS_INCLUDE=""                    # Regex on project:<name> for heartbeats (include).
WAKA_HEARTBEATS_EXCLUDE=""                    # Regex on project:<name> for heartbeats (exclude).
WAKA_HEARTBEATS_CACHE_TTL="15m"               # Cache duration for heartbeats.
WAKA_LOG_LEVEL="info"                         # Log level: debug, info, warn, error.
WAKA_LOG_FORMAT="logfmt"                      # Log format: logfmt or json.
```

## Docker

```shell
docker run -p 9212:9212 -e WAKA_API_KEY="YOUR_API_KEY" thib3113/wakatime-exporter:latest
```

Or use Docker Compose:

```shell
# Linux & Darwin
WAKA_API_KEY="YOUR_API_KEY" docker compose up
```

```powershell
# Windows
$env:WAKA_API_KEY="YOUR_API_KEY"; docker compose up
```

## Authentication

`--wakatime.api-key` (`WAKA_API_KEY`) accepts either a classic Wakatime API key
(Basic auth, e.g. for self-hosted [wakapi](https://github.com/muety/wakapi))
or a WakaTime OAuth access token (`waka_tok_...`, `Authorization: Bearer`).

`--wakatime.auth-mode` (`WAKA_AUTH_MODE`, default `auto`) selects the scheme:
`auto` uses `Bearer` for `waka_tok_` tokens and `Basic` otherwise;
`basic` / `bearer` force the mode (useful for testing and unprefixed tokens).

For personal use (your own account), grab a ready-made OAuth token at
https://wakatime.com/oauth ("Your secret OAuth token") — no app to register.
It can read your email, summaries, stats, goals, organizations, and coding
activity, so keep it secret and prefer it over your API key, which is write-only.

To get an OAuth token for other users, register an app at https://wakatime.com/apps
(see https://wakatime.com/developers#authentication) with at least the scopes
required by the collectors you enable:

| Collector(s)      | Required OAuth scope(s)          |
|-------------------|----------------------------------|
| `summary`, `insights` | `read_summaries`             |
| `all-time`, `stats` | `read_stats`                   |
| `projects`        | `read_stats.projects`            |
| `machines`        | `read_stats.machines`            |
| `user-agents`     | `read_stats.editors`             |
| `goal`            | `read_goals`                     |
| `leader`, `private-leaderboards` | `read_private_leaderboards` |
| `ai`, `ai-breakdown`, `external-durations`, `heartbeats` | `read_heartbeats` |
| `orgs`            | `read_orgs`                      |

Note that the `ai`, `ai-breakdown`, `external-durations` and `heartbeats`
collectors need the broader `read_heartbeats` scope.
`insights`, `heartbeats`, `orgs` and `private-leaderboards` are disabled by default.
Disable any collector you don't need with `--collector.<name>=false`
(or `WAKA_COLLECTOR_<NAME>=false`) to follow the least-privilege principle.

## AI collector

The `ai` collector (enabled by default, endpoint `GET /users/{user}/durations?date=today`)
exposes AI coding statistics:

| Metric                                    | Type    | Labels                  | Description                                              |
|-------------------------------------------|---------|-------------------------|----------------------------------------------------------|
| `wakatime_ai_lines_total`                 | counter | `origin` (`ai`/`human`), `kind` (`additions`/`deletions`) | Lines written by AI vs. typed by a human |
| `wakatime_ai_tokens_total`                | counter | `kind` (`input`/`output`) | AI tokens consumed / produced                          |
| `wakatime_ai_cost_usd_total`              | counter | `model`                 | Estimated AI cost in USD according to WakaTime pricing   |
| `wakatime_ai_sessions_total`              | counter | –                       | Sum of per-duration AI sessions (may double-count overlapping sessions)                                     |
| `wakatime_ai_prompt_events_total`         | counter | –                       | Prompts sent to AI tools                                 |
| `wakatime_ai_prompt_length_chars_sum`     | counter | –                       | Sum of prompt characters                                 |
| `wakatime_ai_prompt_length_avg_chars`     | gauge   | –                       | Mean prompt length in characters                         |
| `wakatime_durations_seconds_total`        | counter | `name`                  | Coded seconds per project, summed from durations (same label as `wakatime_project_seconds_total`) |

It requires the OAuth scope `read_heartbeats` (see [Authentication](#authentication))
and returns no data (`ErrNoData`, never a fatal error) when the endpoint has nothing to report.
Disable it with `--collector.ai=false` or `WAKA_COLLECTOR_AI=false`.

## Stats collector

The `stats` collector (enabled by default, endpoint `GET /users/{user}/stats/{range}`)
exposes totals for a configurable range. The default is `all_time`, i.e. the same
lifetime data the WakaTime UI shows: a project at 120 h is exported as
`wakatime_stats_project_seconds{range="all_time",name="..."} 432000`.

Unlike the `summary` and `ai` collectors (which query *today* and therefore restart
every day), these metrics are gauges whose value depends on `--stats.range`.
They are cached for `--stats.cache-ttl` (default `1h`) to avoid one API call per
scrape. When WakaTime reports stale stats (`is_up_to_date=false`), the exporter
still emits the last known values with `wakatime_stats_up_to_date{range="..."} 0`
and retries after one minute.

| Metric                                      | Type  | Labels                       | Description |
|---------------------------------------------|-------|------------------------------|-------------|
| `wakatime_stats_seconds`                    | gauge | `range`                      | Total seconds, excluding `Other` |
| `wakatime_stats_seconds_including_other`    | gauge | `range`                      | Total seconds, including `Other` |
| `wakatime_stats_daily_average_seconds`      | gauge | `range`                      | Average seconds per day, excluding `Other` |
| `wakatime_stats_daily_average_seconds_including_other` | gauge | `range`         | Average seconds per day, including `Other` |
| `wakatime_stats_best_day_seconds`           | gauge | `range`, `date`              | Seconds on the best day |
| `wakatime_stats_days`                       | gauge | `range`, `kind`              | Days including/minus holidays, or holidays |
| `wakatime_stats_percent_calculated`         | gauge | `range`                      | Background calculation progress |
| `wakatime_stats_is_including_today`         | gauge | `range`                      | Whether today is included |
| `wakatime_stats_is_cached`                  | gauge | `range`                      | Whether the API served cached stats |
| `wakatime_stats_timeout_minutes`            | gauge | `range`                      | Keystroke timeout used for the stats |
| `wakatime_stats_writes_only`                | gauge | `range`                      | Whether writes-only mode was used |
| `wakatime_stats_project_seconds`            | gauge | `range`, `name`              | Seconds per project |
| `wakatime_stats_language_seconds`           | gauge | `range`, `name`              | Seconds per language |
| `wakatime_stats_editor_seconds`             | gauge | `range`, `name`              | Seconds per editor |
| `wakatime_stats_category_seconds`           | gauge | `range`, `name`              | Seconds per category |
| `wakatime_stats_operating_system_seconds`   | gauge | `range`, `name`              | Seconds per operating system |
| `wakatime_stats_machine_seconds`            | gauge | `range`, `id`, `name`        | Seconds per machine |
| `wakatime_stats_dependency_seconds`         | gauge | `range`, `name`              | Seconds per dependency |
| `wakatime_stats_ai_tokens`                  | gauge | `range`, `kind`              | AI tokens used (`input`/`output`) |
| `wakatime_stats_ai_cost_usd`                | gauge | `range`, `model`             | Estimated AI cost in USD according to WakaTime pricing |
| `wakatime_stats_ai_lines`                   | gauge | `range`, `origin`, `kind`    | Lines attributed to AI vs human |
| `wakatime_stats_ai_model_lines`             | gauge | `range`, `model`             | Lines added/removed by each AI model |
| `wakatime_stats_ai_sessions`                | gauge | `range`                      | Distinct AI sessions for the configured range |
| `wakatime_stats_ai_prompt_events`           | gauge | `range`                      | Prompts sent to AI tools |
| `wakatime_stats_project_ai_tokens`          | gauge | `range`, `name`, `kind`      | AI tokens per project |
| `wakatime_stats_project_ai_cost_usd`        | gauge | `range`, `name`, `model`     | Estimated AI cost per project in USD according to WakaTime pricing |
| `wakatime_stats_project_ai_lines`           | gauge | `range`, `name`, `origin`, `kind` | Lines attributed to AI vs human per project |
| `wakatime_stats_project_ai_sessions`        | gauge | `range`, `name`              | AI sessions per project |
| `wakatime_stats_project_ai_prompt_events`   | gauge | `range`, `name`              | AI prompts per project |
| `wakatime_stats_up_to_date`                 | gauge | `range`                      | `1` when WakaTime reports fresh stats |

Per-project AI metrics are only emitted when the project has at least one
non-zero AI field.

`ai_model_costs` are WakaTime cost estimates; use `ai_model_lines` (or tokens)
for usage.

`wakatime_durations_seconds_total` replaces the former
`wakatime_ai_seconds_total`; it still measures all durations, not only
AI-attributed time.

`--stats.range` accepts `all_time`, `last_7_days`, `last_30_days`,
`last_6_months`, `last_year`, a year (`YYYY`), or a month (`YYYY-MM`).
`--stats.top` (default `50`) and `--stats.min-seconds` (default `60`) limit
cardinality; set both to `0` to export every entry. `--stats.include` /
`--stats.exclude` filter `dimension:name` (e.g. `language:Other`) before these
limits are applied. `all-time` emits no metric when WakaTime reports
`is_up_to_date=false`; `stats` emits the last known values with
`wakatime_stats_up_to_date{...} 0`.
Disable the collector with `--collector.stats=false` or `WAKA_COLLECTOR_STATS=false`.

## Additional collectors

All of these are optional and can be disabled with `--collector.<name>=false`
or `WAKA_COLLECTOR_<NAME>=false`. Responses are cached to keep API usage low,
and the exporter limits itself to 8 concurrent WakaTime requests.
`insights`, `heartbeats`, `orgs` and `private-leaderboards` are disabled by
default.

### AI breakdown (`ai-breakdown`)

Queries `/durations?date=today&slice_by=<dimension>` once per configured
dimension and exposes AI usage for that dimension.

| Metric | Labels | Description |
|--------|--------|-------------|
| `wakatime_ai_breakdown_tokens` | `slice`, `name`, `kind` | AI tokens (`input`/`output`) |
| `wakatime_ai_breakdown_cost_usd` | `slice`, `name`, `model` | Estimated AI cost in USD according to WakaTime pricing |
| `wakatime_ai_breakdown_lines` | `slice`, `name`, `origin`, `kind` | Lines attributed to AI vs human |
| `wakatime_ai_breakdown_sessions` | `slice`, `name` | Sum of per-duration AI sessions (may double-count overlapping sessions) |
| `wakatime_ai_breakdown_prompt_events` | `slice`, `name` | Prompts sent to AI tools |
| `wakatime_ai_breakdown_prompt_length_chars_sum` | `slice`, `name` | Sum of prompt characters |

`--ai.breakdown` defaults to `language`; accepted values are `project`,
`entity`, `language`, `dependencies`, `os`, `editor`, `category`, `machine`.
Each dimension is a separate API call, cached for `--ai.cache-ttl` (default
`15m`). `--ai.top` (default `50`) and `--ai.min-tokens` (default `1000`) limit
cardinality. `--ai.include` / `--ai.exclude` filter `<slice>:<name>` (e.g.
`language:Other`). Avoid `entity` and `dependencies` unless you really need
them: their label values are file paths / dependency lists and can explode
cardinality.

### Inventory (`user-agents`, `projects`, `machines`)

| Metric | Labels |
|--------|--------|
| `wakatime_user_agent_info` | `id`, `editor`, `version`, `os`, `ai_model`, `ai_model_version`, `ai_model_complexity`, `browser_extension`, `desktop_app` |
| `wakatime_user_agent_last_seen_timestamp` | `id`, `editor` |
| `wakatime_user_agent_created_timestamp` | `id`, `editor` |
| `wakatime_user_agents_available` | – |
| `wakatime_user_agents_total` | – |
| `wakatime_project_info` | `name`, `id`, `repository`, `has_public_url` |
| `wakatime_project_first_heartbeat_timestamp` | `name` |
| `wakatime_project_last_heartbeat_timestamp` | `name` |
| `wakatime_project_created_timestamp` | `name` |
| `wakatime_projects_available` | – |
| `wakatime_projects_total` | – |
| `wakatime_machine_info` | `id`, `name`, `ip`, `timezone` |
| `wakatime_machine_last_seen_timestamp` | `id`, `name` |
| `wakatime_machine_created_timestamp` | `id`, `name` |
| `wakatime_machines_available` | – |
| `wakatime_machines_total` | – |

These are the fastest way to answer “is my IDE still sending?” and “when did
project X last report?”. They are cached for `--inventory.cache-ttl` (default
`1h`). `--projects.filter` accepts a regex on project names. `--inventory.top`
keeps the N most recent entries, and `--inventory.include` /
`--inventory.exclude` filter `editor:`, `project:` and `machine:` values.
`*_available` is the raw API count before filters/top; `*_total` is the count
actually exported after filters/top.

### Insights (`insights`, disabled by default)

| Metric | Labels | Description |
|--------|--------|-------------|
| `wakatime_insight_seconds` | `range`, `insight`, `name` | Seconds for named insights (projects, languages, editors, categories, machines, operating systems) |
| `wakatime_insight_weekday_seconds` | `range`, `weekday` | Seconds per weekday |
| `wakatime_insight_day_seconds` | `range`, `insight`, `date` | Seconds per day for `days` / `ai_days` |
| `wakatime_insight_best_day_seconds` | `range`, `date` | Seconds on the best day |
| `wakatime_insight_daily_average_seconds` | `range` | Average seconds per day |
| `wakatime_insight_ai_days_ratio` | `range`, `date` | AI line-change percentage per day |
| `wakatime_insight_up_to_date` | `range`, `insight` | `1` when WakaTime reports fresh data |

`--insights.types` defaults to `weekdays,best_day,daily_average,ai_days`;
`--insights.range` defaults to `all_time`. The parser is tolerant: if WakaTime
changes a payload shape, that insight is skipped instead of failing the scrape.
The `stats` insight type is accepted but not exported here (use the `stats`
collector). `--insights.top` (default `50`) and `--insights.min-seconds`
(default `60`) bound named/day insights; set both to `0` for every entry.
`days` and `ai_days` emit one series per date, so lower `--insights.top` (or
omit those types) if cardinality matters.

The Insights API requires a paid WakaTime plan; on a free plan it returns HTTP
402. Enable the collector (`--collector.insights`) only if your plan includes
insights.

### External durations (`external-durations`)

| Metric | Labels |
|--------|--------|
| `wakatime_external_seconds` | `provider`, `category`, `project`, `language` |
| `wakatime_external_durations` | `provider`, `category`, `project`, `language` |
| `wakatime_external_total_seconds` | – |

Use it to separate meetings, calendar events and external AI coding from IDE
coding. `--external-durations.date` defaults to `today`. `--external-durations.include`
/ `--external-durations.exclude` filter `project:`, `provider:`, `category:`
and `language:` values. `wakatime_external_total_seconds` is computed after
those filters.

### Private leaderboards (`private-leaderboards`, disabled by default)

| Metric | Labels |
|--------|--------|
| `wakatime_private_leaderboard_info` | `board`, `name`, `time_range`, `members_count` |
| `wakatime_private_leaderboard_rank` | `board`, `user`, `board_type` |
| `wakatime_private_leaderboard_ai_spend_usd` | `board`, `user`, `board_type` |
| `wakatime_private_leaderboards_available` | – |
| `wakatime_private_leaderboards_total` | – |

`--private-leaderboards.boards` filters by ID (empty = all),
`--private-leaderboards.types` defaults to `time,ai,spend`, and
`--private-leaderboards.max-pages` defaults to `1` and applies to the board
list and to each leaderboard. `available` is the count before the board filter,
`total` is the count exported after it. `ai_spend_usd` is only emitted for
`ai` and `spend` board types.

### Organizations (`orgs`, disabled by default)

| Metric | Labels |
|--------|--------|
| `wakatime_org_info` | `id`, `name`, `timezone`, `default_project_privacy` |
| `wakatime_org_people_count` | `id`, `name` |
| `wakatime_org_invited_people_count` | `id`, `name` |
| `wakatime_org_timeout_minutes` | `id`, `name` |
| `wakatime_org_writes_only` | `id`, `name` |
| `wakatime_orgs_available` | – |
| `wakatime_orgs_total` | – |

Only the organization list is exported; dashboards, members and org summaries
are not (yet). `--orgs.ids` filters by organization ID (empty = all).
`available` is the count before the ID filter, `total` is the count exported
after it.

### Heartbeats (`heartbeats`, disabled by default)

| Metric | Labels |
|--------|--------|
| `wakatime_heartbeats` | `project` |
| `wakatime_heartbeats_write` | `project` |
| `wakatime_heartbeats_ai_tokens` | `project`, `kind` |
| `wakatime_heartbeats_ai_sessions` | `project` |
| `wakatime_heartbeats_total` | – |
| `wakatime_heartbeats_exported_total` | – |

Aggregated by project only: no file, branch or line labels (cardinality and
privacy). `--heartbeats.top` (default `50`) and `--heartbeats.min-count`
(default `1`) limit cardinality. `--heartbeats.include` /
`--heartbeats.exclude` filter `project:<name>`. `heartbeats_total` is the raw
count before filters/top; `heartbeats_exported_total` is the count after them.

### Limiting what you expose

Four complementary levels:

1. **Collector off** — `--collector.<name>=false` or
   `WAKA_COLLECTOR_<NAME>=false`: no API call, no series.
2. **Per-scrape filter** — `GET /metrics?collect[]=summary&collect[]=stats`:
   only the requested collectors run. This is what Prometheus `params:` uses;
   combine it with separate scrape jobs and `scrape_interval` (fast vs slow
   collectors). At most 32 distinct `collect[]` filter sets are cached;
   additional sets are rejected with HTTP 429.
3. **Per-collector limits and filters** — `--stats.top`,
   `--stats.min-seconds`, `--inventory.top`, `--ai.top`, `--ai.min-tokens`,
   `--insights.top`, `--insights.min-seconds`, `--heartbeats.top`, plus the
   composite include/exclude filters below.
4. **Prometheus** — arbitrary drops via `metric_relabel_configs` (on
   `__name__` or any label value) and `write_relabel_configs` before
   remote-writing to Thanos/Mimir.

#### Composite filters (env-friendly)

The high-cardinality collectors accept `--<collector>.include` and
`--<collector>.exclude` regexes matched against `dimension:name`:

| Collector | Values matched |
|-----------|----------------|
| `stats` | `project:<name>`, `language:<name>`, `editor:<name>`, `category:<name>`, `os:<name>`, `machine:<name>`, `dependency:<name>` |
| `ai-breakdown` | `<slice>:<name>` (e.g. `language:TypeScript`) |
| `user-agents`, `projects`, `machines` | `editor:<editor>`, `project:<name>`, `machine:<name>` |
| `heartbeats` | `project:<name>` |
| `external-durations` | `project:<name>`, `provider:<provider>`, `category:<category>`, `language:<name>` |

`include` requires at least one value to match; `exclude` wins if any value
matches. Filters run **before** `top`/`min-*` limits, and an invalid regex
fails at startup. Every operational flag has an environment variable
equivalent (except `--help` and `--version`):

```shell
# stats: drop test/tmp projects and the "Other" language
WAKA_STATS_EXCLUDE='^(project:(test|tmp).*|language:Other)$'
WAKA_STATS_TOP=20
WAKA_STATS_MIN_SECONDS=300

# inventory: keep the 50 most recent machines/user agents/projects
WAKA_INVENTORY_TOP=50
WAKA_INVENTORY_EXCLUDE='^editor:emacs$'

# AI by language: ignore Other/Text
WAKA_AI_EXCLUDE='^language:(Other|Text)$'

# heartbeats / external durations
WAKA_HEARTBEATS_EXCLUDE='^project:(test|tmp)$'
WAKA_EXTERNAL_DURATIONS_INCLUDE='^provider:google_calendar$'
```

`--projects.filter` remains a simpler bare-name include for the projects
collector; the inventory filter can additionally target editors and machines.

To drop a series regardless of the collector logic, use Prometheus
`metric_relabel_configs` (`keep`/`drop` on `__name__` or label values).

Caches are **not** served stale on API errors: if a refresh fails, the
collector emits no value until the next successful fetch.

Some collectors issue multiple sequential requests per refresh (AI breakdown
dimensions, insights types, leaderboard pages). There is no scrape-context
propagation; `--wakatime.timeout` bounds each request and the 30 s handler
timeout bounds the HTTP response, but in-flight work may continue until those
requests finish.

HTTP 429 responses set a global backoff from `Retry-After` (1 s to 5 min);
scrapes fail fast until it expires instead of hammering WakaTime.

### Exposure

The exporter does not authenticate `/metrics`. Machine IPs, repository URLs,
usernames, organization names and editor versions may be exposed. Expose it
only to Prometheus (private network / NetworkPolicy) and protect the endpoint
with a reverse proxy (auth/mTLS) if needed.

## Compatibility

_wakatime_exporter_ is designed to work with [Wakatime](https://wakatime.com/).
However, you can additionally use any other application
that is compliant with the [Wakatime API](https://wakatime.com/developers),
e.g. [wakapi](https://github.com/muety/wakapi),
via `--wakatime.scrape-uri` or `WAKA_SCRAPE_URI`.
If said application only implements portions of the Wakatime API,
you can disable collectors for any non-compliant or non-existent endpoints
using parameters or environment variables as described in [usage](#usage).

## Credits

- Original author: [Jacob Colvin (MacroPower)](https://github.com/MacroPower) —
  created `wakatime_exporter` and maintained it up to v0.1.0. Thank you!
- Current maintainer: [thib3113](https://github.com/thib3113) (OAuth Bearer support,
  `ai`/`stats`/inventory/insights collectors).

The full upstream git history is preserved in this repository, so every commit
from the previous maintainer stays attributed to him (`git log`, `git blame`).
Original `Copyright ... MacroPower` license headers are kept untouched in the
files he wrote; new files carry their own header.

## License

This project was licensed GPL-2.0 from 0.0.0 to 0.0.5.

As of version 0.1.0, parts of this application use code from node_exporter, which is licensed APACHE 2.0.

As of version 0.1.0, this project in its entirety, excluding any packages found in vendor, is licensed APACHE 2.0.
Each file in this project has a license header which identifies the copyright for said file.

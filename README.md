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
> and a new `ai` collector exposing AI coding statistics from the `/durations` endpoint.

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
  --collector.ai                 Enable the ai collector (default: enabled).
  --collector.all-time           Enable the all-time collector (default: enabled).
  --collector.goal               Enable the goal collector (default: enabled).
  --collector.leader             Enable the leader collector (default: enabled).
  --collector.summary            Enable the summary collector (default: enabled).
  --collector.disable-defaults   Set all collectors to disabled by default.
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
WAKA_COLLECTOR_AI="true"                      # Enable the ai collector.
WAKA_COLLECTOR_ALLTIME="true"                 # Enable the all-time collector.
WAKA_COLLECTOR_GOAL="true"                    # Enable the goal collector.
WAKA_COLLECTOR_LEADER="true"                  # Enable the leader collector.
WAKA_COLLECTOR_SUMMARY="true"                 # Enable the summary collector.
```

## Docker

```shell
docker run -p 9212:9212 thib3113/wakatime-exporter:latest --wakatime.api-key="YOUR_API_KEY"
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
| `summary`, `all-time` | `read_summaries`, `read_stats` |
| `goal`            | `read_goals`                     |
| `leader`          | `read_private_leaderboards`      |
| `ai`              | `read_heartbeats`                |

Note that the `ai` collector needs the broader `read_heartbeats` scope.
Disable any collector you don't need with `--collector.<name>=false`
(or `WAKA_COLLECTOR_<NAME>=false`) to follow the least-privilege principle.

## AI collector

The `ai` collector (enabled by default, endpoint `GET /users/{user}/durations?date=today`)
exposes AI coding statistics:

| Metric                                    | Type    | Labels                  | Description                                              |
|-------------------------------------------|---------|-------------------------|----------------------------------------------------------|
| `wakatime_ai_lines_total`                 | counter | `origin` (`ai`/`human`), `kind` (`additions`/`deletions`) | Lines written by AI vs. typed by a human |
| `wakatime_ai_tokens_total`                | counter | `kind` (`input`/`output`) | AI tokens consumed / produced                          |
| `wakatime_ai_cost_usd_total`              | counter | `model`                 | Estimated AI model cost in USD                           |
| `wakatime_ai_sessions_total`              | counter | –                       | Distinct AI sessions                                     |
| `wakatime_ai_prompt_events_total`         | counter | –                       | Prompts sent to AI tools                                 |
| `wakatime_ai_prompt_length_chars_sum`     | counter | –                       | Sum of prompt characters                                 |
| `wakatime_ai_prompt_length_avg_chars`     | gauge   | –                       | Mean prompt length in characters                         |
| `wakatime_ai_seconds_total`               | counter | `name`                  | Coded seconds per project, summed from durations (same label as `wakatime_project_seconds_total`) |

It requires the OAuth scope `read_heartbeats` (see [Authentication](#authentication))
and returns no data (`ErrNoData`, never a fatal error) when the endpoint has nothing to report.
Disable it with `--collector.ai=false` or `WAKA_COLLECTOR_AI=false`.

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
- Current maintainer: [thib3113](https://github.com/thib3113) (OAuth Bearer support, `ai` collector).

The full upstream git history is preserved in this repository, so every commit
from the previous maintainer stays attributed to him (`git log`, `git blame`).
Original `Copyright ... MacroPower` license headers are kept untouched in the
files he wrote; new files carry their own header.

## License

This project was licensed GPL-2.0 from 0.0.0 to 0.0.5.

As of version 0.1.0, parts of this application use code from node_exporter, which is licensed APACHE 2.0.

As of version 0.1.0, this project in its entirety, excluding any packages found in vendor, is licensed APACHE 2.0.
Each file in this project has a license header which identifies the copyright for said file.

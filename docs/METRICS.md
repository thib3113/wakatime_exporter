# Example Metrics

```text
# HELP wakatime_language_seconds_total Total seconds for each language
# TYPE wakatime_language_seconds_total counter
wakatime_language_seconds_total{name="Go"} 3600
wakatime_language_seconds_total{name="Python"} 1800
wakatime_language_seconds_total{name="Markdown"} 900
wakatime_language_seconds_total{name="YAML"} 600
wakatime_language_seconds_total{name="Other"} 300

# HELP wakatime_machine_seconds_total Total seconds for each machine
# TYPE wakatime_machine_seconds_total counter
wakatime_machine_seconds_total{id="m1",name="laptop"} 4200
wakatime_machine_seconds_total{id="m2",name="desktop"} 2400

# HELP wakatime_operating_system_seconds_total Total seconds for each operating system
# TYPE wakatime_operating_system_seconds_total counter
wakatime_operating_system_seconds_total{name="Linux"} 4200
wakatime_operating_system_seconds_total{name="macOS"} 2400

# HELP wakatime_project_seconds_total Total seconds for each project.
# TYPE wakatime_project_seconds_total counter
wakatime_project_seconds_total{name="example-project"} 6600

# HELP wakatime_category_seconds_total Total seconds for each category.
# TYPE wakatime_category_seconds_total counter
wakatime_category_seconds_total{name="Coding"} 6000
wakatime_category_seconds_total{name="Browsing"} 600

# HELP wakatime_editor_seconds_total Total seconds for each editor
# TYPE wakatime_editor_seconds_total counter
wakatime_editor_seconds_total{name="VS Code"} 6600

# HELP wakatime_seconds_total Total seconds.
# TYPE wakatime_seconds_total counter
wakatime_seconds_total 6600

# HELP wakatime_goal_threshold_seconds The goal as set through the wakatime interface.
# TYPE wakatime_goal_threshold_seconds gauge
wakatime_goal_threshold_seconds{delta="day",enabled="true",id="goal-1",ignore_zero_days="1",inverse="false",name="Daily coding goal",snoozed="false",tweeting="false",type="coding"} 3600

# HELP wakatime_goal_progress_seconds Progress towards the goal.
# TYPE wakatime_goal_progress_seconds counter
wakatime_goal_progress_seconds{delta="day",enabled="true",id="goal-1",ignore_zero_days="1",inverse="false",name="Daily coding goal",snoozed="false",tweeting="false",type="coding"} 1800

# HELP wakatime_rank Current rank of the user.
# TYPE wakatime_rank gauge
wakatime_rank 42

# HELP wakatime_cumulative_seconds_total Total seconds (all time).
# TYPE wakatime_cumulative_seconds_total counter
wakatime_cumulative_seconds_total 432000

# HELP wakatime_ai_lines_total Total lines attributed to AI vs human typing, from Wakatime durations.
# TYPE wakatime_ai_lines_total counter
wakatime_ai_lines_total{kind="additions",origin="ai"} 10
wakatime_ai_lines_total{kind="deletions",origin="ai"} 2
wakatime_ai_lines_total{kind="additions",origin="human"} 5
wakatime_ai_lines_total{kind="deletions",origin="human"} 1

# HELP wakatime_ai_tokens_total Total AI tokens used, from Wakatime durations.
# TYPE wakatime_ai_tokens_total counter
wakatime_ai_tokens_total{kind="input"} 120
wakatime_ai_tokens_total{kind="output"} 80

# HELP wakatime_ai_cost_usd_total Estimated AI model cost in USD according to WakaTime pricing.
# TYPE wakatime_ai_cost_usd_total counter
wakatime_ai_cost_usd_total{model="model-alpha"} 0.01

# HELP wakatime_ai_sessions_total Sum of per-duration AI sessions, from Wakatime durations (may double-count overlapping sessions).
# TYPE wakatime_ai_sessions_total counter
wakatime_ai_sessions_total 2

# HELP wakatime_ai_prompt_events_total Total number of prompts sent to AI tools, from Wakatime durations.
# TYPE wakatime_ai_prompt_events_total counter
wakatime_ai_prompt_events_total 4

# HELP wakatime_ai_prompt_length_chars_sum Sum of user prompt characters typed to AI, from Wakatime durations.
# TYPE wakatime_ai_prompt_length_chars_sum counter
wakatime_ai_prompt_length_chars_sum 120

# HELP wakatime_ai_prompt_length_avg_chars Mean prompt length in characters (sum / prompt events), from Wakatime durations.
# TYPE wakatime_ai_prompt_length_avg_chars gauge
wakatime_ai_prompt_length_avg_chars 30

# HELP wakatime_durations_seconds_total Total seconds per project, summed from Wakatime durations.
# TYPE wakatime_durations_seconds_total counter
wakatime_durations_seconds_total{name="example-project"} 300

# HELP wakatime_stats_seconds Coding seconds for the configured range, excluding Other.
# TYPE wakatime_stats_seconds gauge
wakatime_stats_seconds{range="all_time"} 4.32e+05

# HELP wakatime_stats_project_seconds Coding seconds per project for the configured range.
# TYPE wakatime_stats_project_seconds gauge
wakatime_stats_project_seconds{name="example-project",range="all_time"} 4.32e+05

# HELP wakatime_stats_language_seconds Coding seconds per language for the configured range.
# TYPE wakatime_stats_language_seconds gauge
wakatime_stats_language_seconds{name="Go",range="all_time"} 3600

# HELP wakatime_stats_project_ai_tokens AI tokens used per project for the configured range.
# TYPE wakatime_stats_project_ai_tokens gauge
wakatime_stats_project_ai_tokens{kind="input",name="example-project",range="all_time"} 1.2e+05
wakatime_stats_project_ai_tokens{kind="output",name="example-project",range="all_time"} 1.5e+04

# HELP wakatime_stats_ai_model_lines Lines added or removed by each AI model for the configured range.
# TYPE wakatime_stats_ai_model_lines gauge
wakatime_stats_ai_model_lines{model="model-alpha",range="all_time"} 1234
wakatime_stats_ai_model_lines{model="model-beta",range="all_time"} 567
wakatime_stats_ai_model_lines{model="model-gamma",range="all_time"} 234
wakatime_stats_ai_model_lines{model="model-delta",range="all_time"} 89
wakatime_stats_ai_model_lines{model="model-epsilon",range="all_time"} 12

# HELP wakatime_stats_ai_tokens AI tokens used for the configured range.
# TYPE wakatime_stats_ai_tokens gauge
wakatime_stats_ai_tokens{kind="input",range="all_time"} 1.2e+05
wakatime_stats_ai_tokens{kind="output",range="all_time"} 1.5e+04

# HELP wakatime_stats_ai_cost_usd Estimated AI cost in USD according to WakaTime pricing for the configured range.
# TYPE wakatime_stats_ai_cost_usd gauge
wakatime_stats_ai_cost_usd{model="model-alpha",range="all_time"} 0.01
wakatime_stats_ai_cost_usd{model="model-beta",range="all_time"} 0.005

# HELP wakatime_stats_ai_lines Lines attributed to AI vs human for the configured range.
# TYPE wakatime_stats_ai_lines gauge
wakatime_stats_ai_lines{kind="additions",origin="ai",range="all_time"} 10
wakatime_stats_ai_lines{kind="deletions",origin="ai",range="all_time"} 2
wakatime_stats_ai_lines{kind="additions",origin="human",range="all_time"} 5
wakatime_stats_ai_lines{kind="deletions",origin="human",range="all_time"} 1

# HELP wakatime_stats_ai_sessions Distinct AI sessions for the configured range.
# TYPE wakatime_stats_ai_sessions gauge
wakatime_stats_ai_sessions{range="all_time"} 2

# HELP wakatime_stats_ai_prompt_events Prompts sent to AI tools for the configured range.
# TYPE wakatime_stats_ai_prompt_events gauge
wakatime_stats_ai_prompt_events{range="all_time"} 4

# HELP wakatime_stats_best_day_seconds Coding seconds on the best day for the configured range.
# TYPE wakatime_stats_best_day_seconds gauge
wakatime_stats_best_day_seconds{date="2026-01-15",range="all_time"} 36000

# HELP wakatime_stats_days Number of days in the configured range.
# TYPE wakatime_stats_days gauge
wakatime_stats_days{kind="including_holidays",range="all_time"} 30
wakatime_stats_days{kind="minus_holidays",range="all_time"} 20
wakatime_stats_days{kind="holidays",range="all_time"} 10

# HELP wakatime_stats_up_to_date Whether WakaTime reports the stats for the configured range as up to date.
# TYPE wakatime_stats_up_to_date gauge
wakatime_stats_up_to_date{range="all_time"} 1

# HELP wakatime_ai_breakdown_tokens AI tokens used per breakdown dimension, from Wakatime durations.
# TYPE wakatime_ai_breakdown_tokens gauge
wakatime_ai_breakdown_tokens{kind="input",name="Go",slice="language"} 1.2e+05
wakatime_ai_breakdown_tokens{kind="output",name="Go",slice="language"} 1.5e+04

# HELP wakatime_ai_breakdown_cost_usd Estimated AI cost in USD according to WakaTime pricing per breakdown dimension.
# TYPE wakatime_ai_breakdown_cost_usd gauge
wakatime_ai_breakdown_cost_usd{model="model-alpha",name="Go",slice="language"} 0.01

# HELP wakatime_ai_breakdown_lines Lines attributed to AI vs human per breakdown dimension, from Wakatime durations.
# TYPE wakatime_ai_breakdown_lines gauge
wakatime_ai_breakdown_lines{kind="additions",name="Go",origin="ai",slice="language"} 10
wakatime_ai_breakdown_lines{kind="deletions",name="Go",origin="ai",slice="language"} 2
wakatime_ai_breakdown_lines{kind="additions",name="Go",origin="human",slice="language"} 5
wakatime_ai_breakdown_lines{kind="deletions",name="Go",origin="human",slice="language"} 1

# HELP wakatime_ai_breakdown_sessions Sum of per-duration AI sessions per breakdown dimension (may double-count overlapping sessions).
# TYPE wakatime_ai_breakdown_sessions gauge
wakatime_ai_breakdown_sessions{name="Go",slice="language"} 2

# HELP wakatime_ai_breakdown_prompt_events Prompts sent to AI tools per breakdown dimension, from Wakatime durations.
# TYPE wakatime_ai_breakdown_prompt_events gauge
wakatime_ai_breakdown_prompt_events{name="Go",slice="language"} 4

# HELP wakatime_ai_breakdown_prompt_length_chars_sum Sum of prompt characters per breakdown dimension, from Wakatime durations.
# TYPE wakatime_ai_breakdown_prompt_length_chars_sum gauge
wakatime_ai_breakdown_prompt_length_chars_sum{name="Go",slice="language"} 120

# HELP wakatime_user_agent_info WakaTime user agents (IDE plugins) that have sent data.
# TYPE wakatime_user_agent_info gauge
wakatime_user_agent_info{ai_model="model-alpha",ai_model_complexity="high",ai_model_version="4",browser_extension="false",desktop_app="true",editor="VS Code",id="ua1",os="Linux",version="1.2.3"} 1

# HELP wakatime_user_agents_available Number of WakaTime user agents before filters and top limits.
# TYPE wakatime_user_agents_available gauge
wakatime_user_agents_available 100

# HELP wakatime_user_agents_total Number of WakaTime user agents exported after filters and top limits.
# TYPE wakatime_user_agents_total gauge
wakatime_user_agents_total 50

# HELP wakatime_project_last_heartbeat_timestamp Unix timestamp of the last heartbeat for the project.
# TYPE wakatime_project_last_heartbeat_timestamp gauge
wakatime_project_last_heartbeat_timestamp{name="example-project"} 1.7357e+09

# HELP wakatime_projects_available Number of WakaTime projects before filters and top limits.
# TYPE wakatime_projects_available gauge
wakatime_projects_available 5

# HELP wakatime_projects_total Number of WakaTime projects exported after filters and top limits.
# TYPE wakatime_projects_total gauge
wakatime_projects_total 3

# HELP wakatime_machine_info WakaTime machine metadata.
# TYPE wakatime_machine_info gauge
wakatime_machine_info{id="m1",ip="10.0.0.1",name="laptop",timezone="Europe/Paris"} 1

# HELP wakatime_machines_available Number of WakaTime machines before filters and top limits.
# TYPE wakatime_machines_available gauge
wakatime_machines_available 20

# HELP wakatime_machines_total Number of WakaTime machines exported after filters and top limits.
# TYPE wakatime_machines_total gauge
wakatime_machines_total 10

# HELP wakatime_insight_weekday_seconds Coding seconds per weekday.
# TYPE wakatime_insight_weekday_seconds gauge
wakatime_insight_weekday_seconds{range="all_time",weekday="Monday"} 3600

# HELP wakatime_insight_day_seconds Coding seconds per day for day-based insights.
# TYPE wakatime_insight_day_seconds gauge
wakatime_insight_day_seconds{date="2026-01-15",insight="ai_days",range="all_time"} 3600

# HELP wakatime_insight_best_day_seconds Coding seconds on the best day.
# TYPE wakatime_insight_best_day_seconds gauge
wakatime_insight_best_day_seconds{date="2026-01-15",range="all_time"} 7200

# HELP wakatime_insight_daily_average_seconds Average coding seconds per day.
# TYPE wakatime_insight_daily_average_seconds gauge
wakatime_insight_daily_average_seconds{range="all_time"} 1800

# HELP wakatime_insight_ai_days_ratio AI line-change percentage per day from the ai_days insight.
# TYPE wakatime_insight_ai_days_ratio gauge
wakatime_insight_ai_days_ratio{date="2026-01-15",range="all_time"} 42.5

# HELP wakatime_insight_up_to_date Whether WakaTime reports the insight as up to date.
# TYPE wakatime_insight_up_to_date gauge
wakatime_insight_up_to_date{insight="weekdays",range="all_time"} 1

# HELP wakatime_external_seconds Coding seconds from external durations (calendar, meetings, external apps).
# TYPE wakatime_external_seconds gauge
wakatime_external_seconds{category="communicating",language="",project="example-project",provider="google_calendar"} 900

# HELP wakatime_external_total_seconds Total seconds from external durations after filters and limits.
# TYPE wakatime_external_total_seconds gauge
wakatime_external_total_seconds 900

# HELP wakatime_heartbeats Number of heartbeats per project.
# TYPE wakatime_heartbeats gauge
wakatime_heartbeats{project="example-project"} 2

# HELP wakatime_heartbeats_total Total number of heartbeats before filters and top limits.
# TYPE wakatime_heartbeats_total gauge
wakatime_heartbeats_total 200

# HELP wakatime_heartbeats_exported_total Number of heartbeats exported after filters and top limits.
# TYPE wakatime_heartbeats_exported_total gauge
wakatime_heartbeats_exported_total 120

# HELP wakatime_private_leaderboard_rank Rank of a user in a private leaderboard.
# TYPE wakatime_private_leaderboard_rank gauge
wakatime_private_leaderboard_rank{board="b1",board_type="ai",user="alice"} 1

# HELP wakatime_private_leaderboards_available Number of private leaderboards before filters.
# TYPE wakatime_private_leaderboards_available gauge
wakatime_private_leaderboards_available 1

# HELP wakatime_private_leaderboards_total Number of private leaderboards exported after filters.
# TYPE wakatime_private_leaderboards_total gauge
wakatime_private_leaderboards_total 1

# HELP wakatime_org_people_count Number of people in the organization.
# TYPE wakatime_org_people_count gauge
wakatime_org_people_count{id="o1",name="Acme"} 10

# HELP wakatime_orgs_available Number of WakaTime organizations before filters.
# TYPE wakatime_orgs_available gauge
wakatime_orgs_available 1

# HELP wakatime_orgs_total Number of WakaTime organizations exported after filters.
# TYPE wakatime_orgs_total gauge
wakatime_orgs_total 1
```

Additionally, the exporter exports its own metrics, and metrics for each collector (`[ai, ai-breakdown, all-time, external-durations, goal, heartbeats, insights, leader, machines, orgs, private-leaderboards, projects, stats, summary, user-agents]`).

> The `ai` collector queries `GET /users/{user}/durations?date=today` and requires `read_heartbeats`.

> The `ai-breakdown` collector queries `GET /users/{user}/durations?date=today&slice_by=<dimension>` and requires `read_heartbeats`.

> The `stats` collector queries `GET /users/{user}/stats/{range}` and requires `read_stats`; its metrics are gauges because the range is configurable.

> The `insights` collector queries `GET /users/{user}/insights/<type>/<range>` and requires `read_summaries`. It is disabled by default and requires a paid WakaTime plan (HTTP 402 otherwise).

> The `user-agents`, `projects` and `machines` collectors require `read_stats.editors`, `read_stats.projects` and `read_stats.machines` respectively.

> `insights`, `heartbeats`, `orgs` and `private-leaderboards` are disabled by default.

> Inventory `*_available` metrics are raw API counts; `*_total` is the exported count after filters/top. For heartbeats, `heartbeats_total` is raw and `heartbeats_exported_total` is exported.

> `ai_model_costs` are WakaTime cost estimates. Prefer `ai_model_lines` (or tokens) for usage breakdowns.

# Example Metrics

```text
# HELP wakatime_language_seconds_total Total seconds for each language
# TYPE wakatime_language_seconds_total counter
wakatime_language_seconds_total{name="Docker"} 1111.670769
wakatime_language_seconds_total{name="Git Config"} 102.067642
wakatime_language_seconds_total{name="Go"} 8437.99026
wakatime_language_seconds_total{name="JSON"} 121.212324
wakatime_language_seconds_total{name="Jsonnet"} 650.405094
wakatime_language_seconds_total{name="Makefile"} 1926.656756
wakatime_language_seconds_total{name="Markdown"} 1450.281681
wakatime_language_seconds_total{name="Other"} 103.456884
wakatime_language_seconds_total{name="Text"} 5.495999
wakatime_language_seconds_total{name="YAML"} 687.911623

# HELP wakatime_machine_seconds_total Total seconds for each machine
# TYPE wakatime_machine_seconds_total counter
wakatime_machine_seconds_total{id="61940d85-2733-4cbb-b702-adae5ff355d5",name="DesktopPC"} 13917.740195
wakatime_machine_seconds_total{id="8c8e152d-b735-4725-b08c-5f92f35263fb",name="MacbookPro"} 679.408837

# HELP wakatime_operating_system_seconds_total Total seconds for each operating system
# TYPE wakatime_operating_system_seconds_total counter
wakatime_operating_system_seconds_total{name="Mac"} 679.408837
wakatime_operating_system_seconds_total{name="Windows"} 13917.740195

# HELP wakatime_project_seconds_total Total seconds for each project.
# TYPE wakatime_project_seconds_total counter
wakatime_project_seconds_total{name="wakatime_exporter"} 13917.740195

# HELP wakatime_category_seconds_total Total seconds for each category.
# TYPE wakatime_category_seconds_total counter
wakatime_category_seconds_total{name="Coding"} 13917.740195
wakatime_category_seconds_total{name="Browsing"} 679.408837

# HELP wakatime_editor_seconds_total Total seconds for each editor
# TYPE wakatime_editor_seconds_total counter
wakatime_editor_seconds_total{name="VS Code"} 13917.740195

# HELP wakatime_seconds_total Total seconds.
# TYPE wakatime_seconds_total counter
wakatime_seconds_total 14597.149032

# HELP wakatime_goal_info Information about the goal.
# TYPE wakatime_goal_info gauge
wakatime_goal_info{id="69e0febc-923e-4f67-b347-8ed4623823f0",ignore_zero_days="1",is_enabled="1",is_inverse="0",is_snoozed="0",is_tweeting="0",name="Code 1 hr per day in Go"} 1
wakatime_goal_info{id="b642abef-14eb-47c3-ae46-d61a46efe78c",ignore_zero_days="1",is_enabled="1",is_inverse="0",is_snoozed="0",is_tweeting="0",name="Code 1 hr per day in Go, Python"} 1
wakatime_goal_info{id="c3b32c8c-8d4f-44e2-9859-2aa20737f12c",ignore_zero_days="1",is_enabled="1",is_inverse="0",is_snoozed="0",is_tweeting="0",name="Code 1 hr per day"} 1
wakatime_goal_info{id="d8640746-23e6-4553-812a-716efca86020",ignore_zero_days="1",is_enabled="1",is_inverse="0",is_snoozed="0",is_tweeting="0",name="Code 6 hrs per week"} 1

# HELP wakatime_goal_progress_seconds_total Progress towards the goal.
# TYPE wakatime_goal_progress_seconds_total counter
wakatime_goal_progress_seconds_total{delta="day",id="69e0febc-923e-4f67-b347-8ed4623823f0",name="Code 1 hr per day in Go",type="coding"} 3040.641039
wakatime_goal_progress_seconds_total{delta="day",id="b642abef-14eb-47c3-ae46-d61a46efe78c",name="Code 1 hr per day in Go, Python",type="coding"} 3040.641039
wakatime_goal_progress_seconds_total{delta="day",id="c3b32c8c-8d4f-44e2-9859-2aa20737f12c",name="Code 1 hr per day",type="coding"} 3040.641039
wakatime_goal_progress_seconds_total{delta="week",id="d8640746-23e6-4553-812a-716efca86020",name="Code 6 hrs per week",type="coding"} 69270.019606

# HELP wakatime_goal_seconds The goal.
# TYPE wakatime_goal_seconds gauge
wakatime_goal_seconds{delta="day",id="69e0febc-923e-4f67-b347-8ed4623823f0",name="Code 1 hr per day in Go",type="coding"} 3600
wakatime_goal_seconds{delta="day",id="b642abef-14eb-47c3-ae46-d61a46efe78c",name="Code 1 hr per day in Go, Python",type="coding"} 3600
wakatime_goal_seconds{delta="day",id="c3b32c8c-8d4f-44e2-9859-2aa20737f12c",name="Code 1 hr per day",type="coding"} 3600
wakatime_goal_seconds{delta="week",id="d8640746-23e6-4553-812a-716efca86020",name="Code 6 hrs per week",type="coding"} 21600

# HELP wakatime_rank Current rank of the user.
# TYPE wakatime_rank gauge
wakatime_rank 2886

# HELP wakatime_cumulative_seconds_total Total seconds (all time).
# TYPE wakatime_cumulative_seconds_total counter
wakatime_cumulative_seconds_total 219501.379948

# HELP wakatime_ai_lines_total Total lines attributed to AI vs human typing, from Wakatime durations.
# TYPE wakatime_ai_lines_total counter
wakatime_ai_lines_total{kind="additions",origin="ai"} 13
wakatime_ai_lines_total{kind="deletions",origin="ai"} 2
wakatime_ai_lines_total{kind="additions",origin="human"} 5
wakatime_ai_lines_total{kind="deletions",origin="human"} 1

# HELP wakatime_ai_tokens_total Total AI tokens used, from Wakatime durations.
# TYPE wakatime_ai_tokens_total counter
wakatime_ai_tokens_total{kind="input"} 150
wakatime_ai_tokens_total{kind="output"} 210

# HELP wakatime_ai_cost_usd_total Estimated AI model cost in USD, from Wakatime durations.
# TYPE wakatime_ai_cost_usd_total counter
wakatime_ai_cost_usd_total{model="claude"} 0.0133

# HELP wakatime_ai_sessions_total Total distinct AI sessions, from Wakatime durations.
# TYPE wakatime_ai_sessions_total counter
wakatime_ai_sessions_total 2

# HELP wakatime_ai_prompt_events_total Total number of prompts sent to AI tools, from Wakatime durations.
# TYPE wakatime_ai_prompt_events_total counter
wakatime_ai_prompt_events_total 3

# HELP wakatime_ai_prompt_length_chars_sum Sum of user prompt characters typed to AI, from Wakatime durations.
# TYPE wakatime_ai_prompt_length_chars_sum counter
wakatime_ai_prompt_length_chars_sum 100

# HELP wakatime_ai_prompt_length_avg_chars Mean prompt length in characters (sum / prompt events), from Wakatime durations.
# TYPE wakatime_ai_prompt_length_avg_chars gauge
wakatime_ai_prompt_length_avg_chars 33.333333333333336

# HELP wakatime_ai_seconds_total Total seconds per project, summed from Wakatime durations.
# TYPE wakatime_ai_seconds_total counter
wakatime_ai_seconds_total{name="wakatime-exporter"} 200
```

Additionally, wakatime_exporter exports its own metrics, and metrics for each collector (`[ai, alltime, goal, leader, summary]`).

> The `ai` collector queries `GET /users/{user}/durations?date=today` and requires the OAuth scope `read_heartbeats` (broader than the other collectors).

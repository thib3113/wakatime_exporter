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

import "time"

// Config groups the per-collector options.
type Config struct {
	Stats               StatsConfig
	AI                  AIBreakdownConfig
	Insights            InsightsConfig
	Inventory           InventoryConfig
	ExternalDurations   ExternalDurationsConfig
	PrivateLeaderboards PrivateLeaderboardsConfig
	Orgs                OrgsConfig
	Heartbeats          HeartbeatsConfig
}

type FilterConfig struct {
	Include string
	Exclude string
}

type StatsConfig struct {
	Range      string
	CacheTTL   time.Duration
	Top        int
	MinSeconds float64
	Filter     FilterConfig
}

type AIBreakdownConfig struct {
	Dimensions []string
	Top        int
	MinTokens  int64
	CacheTTL   time.Duration
	Filter     FilterConfig
}

type InsightsConfig struct {
	Types      []string
	Range      string
	CacheTTL   time.Duration
	Top        int
	MinSeconds float64
}

type InventoryConfig struct {
	Top            int
	Filter         FilterConfig
	ProjectsFilter string
	CacheTTL       time.Duration
}

type ExternalDurationsConfig struct {
	Date     string
	CacheTTL time.Duration
	Filter   FilterConfig
}

type PrivateLeaderboardsConfig struct {
	Boards   []string
	Types    []string
	MaxPages int
	CacheTTL time.Duration
}

type OrgsConfig struct {
	IDs      []string
	CacheTTL time.Duration
}

type HeartbeatsConfig struct {
	Top      int
	MinCount int
	CacheTTL time.Duration
	Filter   FilterConfig
}

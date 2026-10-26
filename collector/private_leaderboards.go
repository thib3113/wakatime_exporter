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
	"regexp"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	privateLeaderboardsCollectorName = "private-leaderboards"
	privateLeaderboardsEndpoint      = "leaderboards"
)

var (
	privateLeaderboardTypes = map[string]bool{
		"time":   true,
		"manual": true,
		"ai":     true,
		"spend":  true,
	}
	boardIDPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
)

func validBoardID(id string) bool {
	return id != "." && id != ".." && boardIDPattern.MatchString(id)
}

type privateLeaderboardsCollector struct {
	info      *prometheus.Desc
	rank      *prometheus.Desc
	aiSpend   *prometheus.Desc
	available *prometheus.Desc
	total     *prometheus.Desc
	boards    []string
	types     []string
	maxPages  int
	cache     ttlCache[privateLeaderboardSnapshot]
	cacheTTL  time.Duration
	uri       url.URL
	fetchStat func(url.URL, string, url.Values) (io.ReadCloser, error)
	logger    log.Logger
}

type privateLeaderboardSnapshot struct {
	Available int
	Boards    []wakatimePrivateLeaderboard
	Leaders   map[string]map[string][]wakatimePrivateLeaderboardLeader
}

func init() {
	registerCollector(privateLeaderboardsCollectorName, defaultDisabled, NewPrivateLeaderboardsCollector)
}

func NewPrivateLeaderboardsCollector(in CommonInputs, logger log.Logger) (Collector, error) {
	types := in.Config.PrivateLeaderboards.Types
	if len(types) == 0 {
		types = []string{"time", "ai", "spend"}
	}
	for _, boardType := range types {
		if !privateLeaderboardTypes[boardType] {
			return nil, fmt.Errorf("invalid private leaderboard type %q", boardType)
		}
	}

	maxPages := in.Config.PrivateLeaderboards.MaxPages
	if maxPages <= 0 {
		maxPages = 1
	}
	cacheTTL := in.Config.PrivateLeaderboards.CacheTTL
	if cacheTTL <= 0 {
		cacheTTL = time.Hour
	}

	return &privateLeaderboardsCollector{
		info: newMetricDesc(
			"private_leaderboard_info",
			"WakaTime private leaderboard metadata.",
			"board", "name", "time_range", "members_count",
		),
		rank: newMetricDesc(
			"private_leaderboard_rank",
			"Rank of a user in a private leaderboard.",
			"board", "user", "board_type",
		),
		aiSpend: newMetricDesc(
			"private_leaderboard_ai_spend_usd",
			"AI spend of a user in a private leaderboard.",
			"board", "user", "board_type",
		),
		available: newMetricDesc(
			"private_leaderboards_available",
			"Number of private leaderboards before filters.",
		),
		total: newMetricDesc(
			"private_leaderboards_total",
			"Number of private leaderboards exported after filters.",
		),
		boards:    in.Config.PrivateLeaderboards.Boards,
		types:     types,
		maxPages:  maxPages,
		cacheTTL:  cacheTTL,
		uri:       in.URI,
		fetchStat: FetchHTTP(in.HTTPClient, in.Token, in.AuthMode, logger),
		logger:    logger,
	}, nil
}

func (c *privateLeaderboardsCollector) Update(ch chan<- prometheus.Metric) error {
	snapshot, err := c.cache.get(c.cacheTTL, c.fetchSnapshot)
	if err != nil {
		return err
	}

	ch <- prometheus.MustNewConstMetric(c.available, prometheus.GaugeValue, float64(snapshot.Available))
	ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, float64(len(snapshot.Boards)))
	if len(snapshot.Boards) == 0 {
		return nil
	}

	level.Info(c.logger).Log("msg", "Collecting private leaderboards from Wakatime", "boards", len(snapshot.Boards))
	for _, board := range snapshot.Boards {
		ch <- prometheus.MustNewConstMetric(c.info, prometheus.GaugeValue, 1,
			board.ID,
			board.Name,
			board.TimeRange,
			strconv.Itoa(board.MembersCount),
		)
		for _, boardType := range c.types {
			for _, leader := range snapshot.Leaders[board.ID][boardType] {
				user := leader.User.Username
				if user == "" {
					user = leader.User.ID
				}
				if user == "" {
					user = "unknown"
				}
				ch <- prometheus.MustNewConstMetric(c.rank, prometheus.GaugeValue, float64(leader.Rank), board.ID, user, boardType)
				if boardType == "ai" || boardType == "spend" {
					ch <- prometheus.MustNewConstMetric(c.aiSpend, prometheus.GaugeValue, leader.RunningTotal.AIModelTotalCost, board.ID, user, boardType)
				}
			}
		}
	}

	return nil
}

func (c *privateLeaderboardsCollector) fetchSnapshot() (*privateLeaderboardSnapshot, error) {
	boards, err := paginate(c.maxPages, func(page int) ([]wakatimePrivateLeaderboard, int, error) {
		params := url.Values{}
		params.Set("page", strconv.Itoa(page))

		body, fetchErr := c.fetchStat(c.uri, privateLeaderboardsEndpoint, params)
		if fetchErr != nil {
			return nil, 0, fetchErr
		}
		var boardsResponse wakatimePrivateLeaderboards
		unmarshalErr := ReadAndUnmarshal(body, &boardsResponse)
		body.Close()
		if unmarshalErr != nil {
			return nil, 0, unmarshalErr
		}
		return boardsResponse.Data, boardsResponse.TotalPages, nil
	})
	if err != nil {
		return nil, err
	}

	allowed := map[string]bool{}
	for _, board := range c.boards {
		allowed[board] = true
	}

	snapshot := &privateLeaderboardSnapshot{
		Available: len(boards),
		Boards:    make([]wakatimePrivateLeaderboard, 0, len(boards)),
		Leaders:   map[string]map[string][]wakatimePrivateLeaderboardLeader{},
	}
	for _, board := range boards {
		if len(allowed) > 0 && !allowed[board.ID] {
			continue
		}
		if !validBoardID(board.ID) {
			level.Warn(c.logger).Log("msg", "Invalid private leaderboard id", "id", board.ID)
			continue
		}
		snapshot.Boards = append(snapshot.Boards, board)
		snapshot.Leaders[board.ID] = map[string][]wakatimePrivateLeaderboardLeader{}

		for _, boardType := range c.types {
			leaders, err := paginate(c.maxPages, func(page int) ([]wakatimePrivateLeaderboardLeader, int, error) {
				params := url.Values{}
				params.Set("board_type", boardType)
				params.Set("page", strconv.Itoa(page))

				leadersBody, fetchErr := c.fetchStat(c.uri, privateLeaderboardsEndpoint+"/"+board.ID, params)
				if fetchErr != nil {
					return nil, 0, fetchErr
				}
				var leadersResponse wakatimePrivateLeaderboardLeaders
				unmarshalErr := ReadAndUnmarshal(leadersBody, &leadersResponse)
				leadersBody.Close()
				if unmarshalErr != nil {
					return nil, 0, unmarshalErr
				}
				return leadersResponse.Data, leadersResponse.TotalPages, nil
			})
			if err != nil {
				return nil, err
			}
			snapshot.Leaders[board.ID][boardType] = leaders
		}
	}

	return snapshot, nil
}

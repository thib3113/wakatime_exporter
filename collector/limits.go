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
	"sort"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const warnInterval = 5 * time.Minute

var (
	warnMu   sync.Mutex
	warnLast = map[string]time.Time{}
)

type namedValue struct {
	name  string
	value float64
}

func topNamedValues(values map[string]float64, top int, minValue float64) []namedValue {
	all := make([]namedValue, 0, len(values))
	for name, value := range values {
		if minValue > 0 && value < minValue {
			continue
		}
		all = append(all, namedValue{name: name, value: value})
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].value == all[j].value {
			return all[i].name < all[j].name
		}
		return all[i].value > all[j].value
	})
	if top > 0 && len(all) > top {
		all = all[:top]
	}
	return all
}

func trimWarn[T any](logger log.Logger, collector, dimension string, items []T, name func(T) string, seconds func(T) float64, top int, minSeconds float64, filter *nameFilter) []T {
	out, dropped := trimStats(items, dimension, name, seconds, top, minSeconds, filter)
	warnDropped(logger, collector, dimension, len(out)+dropped, len(out))
	return out
}

func warnDropped(logger log.Logger, collector, dimension string, total, kept int) {
	if total <= kept {
		return
	}

	key := collector + "|" + dimension
	now := time.Now()
	warnMu.Lock()
	if last, ok := warnLast[key]; ok && now.Sub(last) < warnInterval {
		warnMu.Unlock()
		return
	}
	warnLast[key] = now
	warnMu.Unlock()

	level.Warn(logger).Log(
		"msg", "Dropping entries due to cardinality limit",
		"collector", collector,
		"dimension", dimension,
		"total", total,
		"kept", kept,
	)
}

func resetWarnState() {
	warnMu.Lock()
	warnLast = map[string]time.Time{}
	warnMu.Unlock()
}

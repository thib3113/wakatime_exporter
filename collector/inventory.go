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

import "sort"

func selectInventory[T any](items []T, dimension string, name func(T) string, timestamp func(T) float64, filter *nameFilter, top int) ([]T, int) {
	out := make([]T, 0, len(items))
	for _, item := range items {
		if !filter.Match(dimensionValue(dimension, name(item))) {
			continue
		}
		out = append(out, item)
	}
	total := len(out)
	sort.SliceStable(out, func(i, j int) bool {
		left, right := timestamp(out[i]), timestamp(out[j])
		if left == right {
			return name(out[i]) < name(out[j])
		}
		return left > right
	})
	if top > 0 && len(out) > top {
		out = out[:top]
	}
	return out, total - len(out)
}

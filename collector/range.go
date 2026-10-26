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
	"regexp"
	"time"
)

var (
	rangeYearPattern  = regexp.MustCompile(`^\d{4}$`)
	rangeMonthPattern = regexp.MustCompile(`^\d{4}-\d{2}$`)
)

func validateRange(value string) error {
	switch value {
	case "all_time", "last_7_days", "last_30_days", "last_6_months", "last_year":
		return nil
	}
	if rangeYearPattern.MatchString(value) {
		return nil
	}
	if rangeMonthPattern.MatchString(value) {
		if _, err := time.Parse("2006-01", value); err != nil {
			return fmt.Errorf("invalid range %q", value)
		}
		return nil
	}
	return fmt.Errorf("invalid range %q", value)
}

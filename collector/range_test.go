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

import "testing"

func TestValidateRange(t *testing.T) {
	for _, value := range []string{
		"all_time", "last_7_days", "last_30_days", "last_6_months", "last_year", "2026", "2026-01",
	} {
		if err := validateRange(value); err != nil {
			t.Fatalf("validateRange(%q): %v", value, err)
		}
	}
	for _, value := range []string{"", "foobar", "2026-13", "20-01", "../etc", "last_2_days"} {
		if err := validateRange(value); err == nil {
			t.Fatalf("validateRange(%q) should fail", value)
		}
	}
}

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

func TestNameFilter(t *testing.T) {
	filter, err := newNameFilter(FilterConfig{
		Include: `^(project:example.*|language:Go)$`,
		Exclude: `^project:example-test$`,
	})
	if err != nil {
		t.Fatalf("newNameFilter: %v", err)
	}

	cases := []struct {
		value string
		want  bool
	}{
		{"project:example-project", true},
		{"project:example-test", false},
		{"language:Go", true},
		{"language:TypeScript", false},
		{"machine:laptop", false},
	}
	for _, tc := range cases {
		if got := filter.Match(tc.value); got != tc.want {
			t.Fatalf("Match(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}

	if filter.MatchAny("project:example-project", "language:Go") != true {
		t.Fatal("MatchAny should match if any value is included")
	}
	if filter.MatchAny("project:example-test", "language:Go") != false {
		t.Fatal("MatchAny should reject if any value is excluded")
	}

	includeProvider, err := newNameFilter(FilterConfig{Include: `^provider:google_calendar$`})
	if err != nil {
		t.Fatalf("newNameFilter: %v", err)
	}
	if !includeProvider.MatchAny("project:x", "provider:google_calendar") {
		t.Fatal("include should match any provided value")
	}
	if includeProvider.MatchAny("project:x", "provider:zoom") {
		t.Fatal("include should not match when no value matches")
	}

	excludeCategory, err := newNameFilter(FilterConfig{Exclude: `^category:communicating$`})
	if err != nil {
		t.Fatalf("newNameFilter: %v", err)
	}
	if excludeCategory.MatchAny("project:x", "category:communicating") {
		t.Fatal("exclude should reject when any value matches")
	}

	empty, err := newNameFilter(FilterConfig{})
	if err != nil {
		t.Fatalf("newNameFilter: %v", err)
	}
	if empty != nil {
		t.Fatal("empty filter should be nil")
	}

	if _, err := newNameFilter(FilterConfig{Include: "("}); err == nil {
		t.Fatal("invalid include regex should fail")
	}
	if _, err := newNameFilter(FilterConfig{Exclude: "["}); err == nil {
		t.Fatal("invalid exclude regex should fail")
	}

	if got := dimensionValue("project", "example"); got != "project:example" {
		t.Fatalf("dimensionValue = %q", got)
	}
}

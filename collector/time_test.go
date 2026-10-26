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
	"testing"
	"time"
)

func TestParseUnixSeconds(t *testing.T) {
	want := float64(time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC).Unix())
	got, ok := parseUnixSeconds("2026-01-15T12:00:00Z")
	if !ok || got != want {
		t.Fatalf("parseUnixSeconds = %v, %v; want %v, true", got, ok, want)
	}
	if _, ok := parseUnixSeconds(""); ok {
		t.Fatal("empty timestamp should be invalid")
	}
	if _, ok := parseUnixSeconds("not-a-date"); ok {
		t.Fatal("invalid timestamp should be invalid")
	}
	if got := parseUnixSecondsOrZero(""); got != 0 {
		t.Fatalf("parseUnixSecondsOrZero = %v, want 0", got)
	}
}

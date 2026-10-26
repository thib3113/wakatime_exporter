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

package main

import "testing"

func TestFilterKey(t *testing.T) {
	got := filterKey([]string{"stats", "summary", "stats", "all-time"})
	want := filterKey([]string{"all-time", "stats", "summary"})
	if got != want {
		t.Fatalf("filterKey = %q, want %q", got, want)
	}
}

func TestFilterKeyNoCollision(t *testing.T) {
	if filterKey([]string{"a,b"}) == filterKey([]string{"a", "b"}) {
		t.Fatal("filterKey must not collide on embedded separators")
	}
}

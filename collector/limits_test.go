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
	"bytes"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestTopNamedValues(t *testing.T) {
	values := map[string]float64{
		"low":      5,
		"high":     100,
		"medium":   50,
		"tie-high": 100,
	}

	got := topNamedValues(values, 2, 10)
	if len(got) != 2 {
		t.Fatalf("got %d values, want 2", len(got))
	}
	if got[0].name != "high" || got[1].name != "tie-high" {
		t.Fatalf("unexpected top order: %#v", got)
	}

	got = topNamedValues(values, 2, 0)
	if len(got) != 2 || got[0].name != "high" || got[1].name != "tie-high" {
		t.Fatalf("unexpected top with min=0: %#v", got)
	}

	if got := topNamedValues(values, 0, 10); len(got) != 3 {
		t.Fatalf("expected 3 values after min filter, got %d", len(got))
	}
}

func TestWarnDropped(t *testing.T) {
	resetWarnState()
	var buf bytes.Buffer
	logger := log.NewLogfmtLogger(&buf)
	warnDropped(logger, "test", "dimension", 10, 5)
	if !strings.Contains(buf.String(), "Dropping entries due to cardinality limit") {
		t.Fatalf("unexpected log output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "total=10") || !strings.Contains(buf.String(), "kept=5") {
		t.Fatalf("missing counts in log output: %q", buf.String())
	}
}

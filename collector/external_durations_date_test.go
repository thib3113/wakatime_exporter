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

	"github.com/go-kit/kit/log"
)

func TestExternalDurationsDateValidation(t *testing.T) {
	for _, date := range []string{"today", "yesterday", "2026-01-15"} {
		if _, err := NewExternalDurationsCollector(CommonInputs{
			Config: Config{ExternalDurations: ExternalDurationsConfig{Date: date}},
		}, log.NewNopLogger()); err != nil {
			t.Fatalf("date %q: %v", date, err)
		}
	}

	if _, err := NewExternalDurationsCollector(CommonInputs{
		Config: Config{ExternalDurations: ExternalDurationsConfig{Date: "2026-13-99"}},
	}, log.NewNopLogger()); err == nil {
		t.Fatal("expected invalid date to be rejected")
	}
}

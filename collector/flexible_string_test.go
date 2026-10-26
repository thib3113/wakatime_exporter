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
	"encoding/json"
	"testing"
)

func TestFlexibleString(t *testing.T) {
	type payload struct {
		Value flexibleString `json:"value"`
	}

	cases := []struct {
		raw     string
		want    string
		wantErr bool
	}{
		{`{"value":"Europe/Paris"}`, "Europe/Paris", false},
		{`{"value":42}`, "42", false},
		{`{"value":true}`, "true", false},
		{`{"value":null}`, "", false},
		{`{"value":{}}`, "", true},
	}
	for _, tc := range cases {
		var decoded payload
		err := json.Unmarshal([]byte(tc.raw), &decoded)
		if tc.wantErr {
			if err == nil {
				t.Fatalf("%s: expected an error", tc.raw)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: %v", tc.raw, err)
		}
		if string(decoded.Value) != tc.want {
			t.Fatalf("%s: got %q, want %q", tc.raw, decoded.Value, tc.want)
		}
	}
}

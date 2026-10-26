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
	"net/http"
	"strings"
	"testing"
)

func TestHTTPStatusError(t *testing.T) {
	cases := []struct {
		status int
		header http.Header
		want   string
	}{
		{401, http.Header{}, "unauthorized"},
		{402, http.Header{}, "payment required"},
		{403, http.Header{}, "forbidden"},
		{429, http.Header{"Retry-After": []string{"30"}}, "retry after 30"},
		{500, http.Header{}, "HTTP status 500"},
	}
	for _, tc := range cases {
		err := httpStatusError(&http.Response{StatusCode: tc.status, Header: tc.header})
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("status %d: got %v, want containing %q", tc.status, err, tc.want)
		}
	}
}

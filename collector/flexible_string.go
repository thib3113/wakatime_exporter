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
	"encoding/json"
	"fmt"
	"strconv"
)

type flexibleString string

func (s *flexibleString) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) {
		*s = ""
		return nil
	}

	var str string
	if json.Unmarshal(data, &str) == nil {
		*s = flexibleString(str)
		return nil
	}

	var number json.Number
	if json.Unmarshal(data, &number) == nil {
		*s = flexibleString(number.String())
		return nil
	}

	var boolean bool
	if json.Unmarshal(data, &boolean) == nil {
		*s = flexibleString(strconv.FormatBool(boolean))
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into string", string(data))
}

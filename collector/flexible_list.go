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
)

// flexibleList accepts either a JSON array or a single JSON object.
// Some WakaTime endpoints document a list but show a single object in their
// examples, so tolerate both shapes.
type flexibleList[T any] []T

func (l *flexibleList[T]) UnmarshalJSON(data []byte) error {
	var compact bytes.Buffer
	if err := json.Compact(&compact, data); err == nil {
		switch compact.String() {
		case "null", "{}":
			*l = nil
			return nil
		}
	}

	var items []T
	if err := json.Unmarshal(data, &items); err == nil {
		*l = items
		return nil
	}

	var item T
	if err := json.Unmarshal(data, &item); err != nil {
		return err
	}
	*l = flexibleList[T]{item}
	return nil
}

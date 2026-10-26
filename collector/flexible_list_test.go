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

func TestFlexibleList(t *testing.T) {
	type item struct {
		Name string `json:"name"`
	}

	var array struct {
		Data flexibleList[item] `json:"data"`
	}
	if err := json.Unmarshal([]byte(`{"data":[{"name":"a"},{"name":"b"}]}`), &array); err != nil {
		t.Fatalf("array: %v", err)
	}
	if len(array.Data) != 2 || array.Data[0].Name != "a" || array.Data[1].Name != "b" {
		t.Fatalf("unexpected array: %#v", array.Data)
	}

	var object struct {
		Data flexibleList[item] `json:"data"`
	}
	if err := json.Unmarshal([]byte(`{"data":{"name":"a"}}`), &object); err != nil {
		t.Fatalf("object: %v", err)
	}
	if len(object.Data) != 1 || object.Data[0].Name != "a" {
		t.Fatalf("unexpected object: %#v", object.Data)
	}

	var empty struct {
		Data flexibleList[item] `json:"data"`
	}
	if err := json.Unmarshal([]byte(`{"data":{}}`), &empty); err != nil {
		t.Fatalf("empty: %v", err)
	}
	if len(empty.Data) != 0 {
		t.Fatalf("expected empty list, got %#v", empty.Data)
	}

	var spacedEmpty struct {
		Data flexibleList[item] `json:"data"`
	}
	if err := json.Unmarshal([]byte(`{"data":{ }}`), &spacedEmpty); err != nil {
		t.Fatalf("spaced empty: %v", err)
	}
	if len(spacedEmpty.Data) != 0 {
		t.Fatalf("expected empty list, got %#v", spacedEmpty.Data)
	}

	var null struct {
		Data flexibleList[item] `json:"data"`
	}
	if err := json.Unmarshal([]byte(`{"data":null}`), &null); err != nil {
		t.Fatalf("null: %v", err)
	}
	if len(null.Data) != 0 {
		t.Fatalf("expected empty list, got %#v", null.Data)
	}
}

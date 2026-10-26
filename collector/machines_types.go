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

type wakatimeMachines struct {
	Data flexibleList[wakatimeMachine] `json:"data"`
}

type wakatimeMachine struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Value      string `json:"value"`
	IP         string `json:"ip"`
	LastSeenAt string `json:"last_seen_at"`
	Timezone   string `json:"timezone"`
	CreatedAt  string `json:"created_at"`
}

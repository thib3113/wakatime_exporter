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

type wakatimeProjects struct {
	Data flexibleList[wakatimeProject] `json:"data"`
}

type wakatimeProject struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Repository       string `json:"repository"`
	Badge            string `json:"badge"`
	Color            string `json:"color"`
	HasPublicURL     bool   `json:"has_public_url"`
	LastHeartbeatAt  string `json:"last_heartbeat_at"`
	FirstHeartbeatAt string `json:"first_heartbeat_at"`
	URL              string `json:"url"`
	URLEncodedName   string `json:"urlencoded_name"`
	CreatedAt        string `json:"created_at"`
}

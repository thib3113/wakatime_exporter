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
	"sync"
	"time"
)

type ttlCache[T any] struct {
	mu    sync.Mutex
	value *T
	until time.Time
}

func (c *ttlCache[T]) get(ttl time.Duration, fetch func() (*T, error)) (*T, error) {
	c.mu.Lock()
	if c.value != nil && time.Now().Before(c.until) {
		value := c.value
		c.mu.Unlock()
		return value, nil
	}
	c.mu.Unlock()

	value, err := fetch()
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.value = value
	c.until = time.Now().Add(ttl)
	c.mu.Unlock()
	return value, nil
}

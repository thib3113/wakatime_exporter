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

type keyedTTLCache[T any] struct {
	mu      sync.Mutex
	entries map[string]keyedCacheEntry[T]
	flights flightGroup
}

type keyedCacheEntry[T any] struct {
	value *T
	until time.Time
}

func (c *keyedTTLCache[T]) get(key string, ttl time.Duration, fetch func() (*T, error)) (*T, error) {
	return c.getWithTTL(key, func() (*T, time.Duration, error) {
		value, err := fetch()
		return value, ttl, err
	})
}

func (c *keyedTTLCache[T]) getWithTTL(key string, fetch func() (*T, time.Duration, error)) (*T, error) {
	c.mu.Lock()
	if entry, ok := c.entries[key]; ok && time.Now().Before(entry.until) {
		value := entry.value
		c.mu.Unlock()
		return value, nil
	}
	c.mu.Unlock()

	return singleflightDo(&c.flights, key, func() (*T, error) {
		value, ttl, err := fetch()
		if err != nil {
			return nil, err
		}

		c.mu.Lock()
		if c.entries == nil {
			c.entries = map[string]keyedCacheEntry[T]{}
		}
		c.entries[key] = keyedCacheEntry[T]{value: value, until: time.Now().Add(ttl)}
		c.mu.Unlock()
		return value, nil
	})
}

type ttlCache[T any] struct {
	cache keyedTTLCache[T]
}

func (c *ttlCache[T]) get(ttl time.Duration, fetch func() (*T, error)) (*T, error) {
	return c.cache.get("value", ttl, fetch)
}

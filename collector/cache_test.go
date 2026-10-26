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
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTTLCache(t *testing.T) {
	var cache ttlCache[int]
	calls := 0
	fetch := func() (*int, error) {
		calls++
		value := calls
		return &value, nil
	}

	first, err := cache.get(time.Hour, fetch)
	if err != nil {
		t.Fatalf("first get: %v", err)
	}
	second, err := cache.get(time.Hour, fetch)
	if err != nil {
		t.Fatalf("second get: %v", err)
	}
	if calls != 1 || *first != 1 || *second != 1 {
		t.Fatalf("cache miss: calls=%d first=%d second=%d", calls, *first, *second)
	}

	var expiring ttlCache[int]
	calls = 0
	if _, err := expiring.get(time.Nanosecond, fetch); err != nil {
		t.Fatalf("expiring first get: %v", err)
	}
	time.Sleep(time.Millisecond)
	if _, err := expiring.get(time.Hour, fetch); err != nil {
		t.Fatalf("expiring second get: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected cache expiry, calls=%d", calls)
	}

	expected := errors.New("boom")
	var failing ttlCache[int]
	if _, err := failing.get(time.Hour, func() (*int, error) { return nil, expected }); !errors.Is(err, expected) {
		t.Fatalf("expected fetch error, got %v", err)
	}
}

func TestTTLCacheConcurrent(t *testing.T) {
	var cache ttlCache[int]
	var wg sync.WaitGroup
	var calls int32
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := cache.get(time.Hour, func() (*int, error) {
				atomic.AddInt32(&calls, 1)
				time.Sleep(10 * time.Millisecond)
				result := 42
				return &result, nil
			})
			if err != nil {
				t.Errorf("get: %v", err)
				return
			}
			if *value != 42 {
				t.Errorf("got %d, want 42", *value)
			}
		}()
	}
	wg.Wait()
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected a single fetch, got %d", got)
	}
}

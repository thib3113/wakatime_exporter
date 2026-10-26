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
	"fmt"
	"sync"
)

type flightGroup struct {
	mu    sync.Mutex
	calls map[string]*flightCall
}

type flightCall struct {
	done  chan struct{}
	value any
	err   error
}

func (g *flightGroup) do(key string, fn func() (any, error)) (value any, err error) {
	g.mu.Lock()
	if call, ok := g.calls[key]; ok {
		g.mu.Unlock()
		<-call.done
		return call.value, call.err
	}
	if g.calls == nil {
		g.calls = map[string]*flightCall{}
	}
	call := &flightCall{done: make(chan struct{})}
	g.calls[key] = call
	g.mu.Unlock()

	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("panic in singleflight: %v", recovered)
		}
		call.value = value
		call.err = err
		g.mu.Lock()
		delete(g.calls, key)
		g.mu.Unlock()
		close(call.done)
	}()

	value, err = fn()
	return value, err
}

func singleflightDo[T any](g *flightGroup, key string, fn func() (T, error)) (T, error) {
	value, err := g.do(key, func() (any, error) {
		return fn()
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return value.(T), nil
}

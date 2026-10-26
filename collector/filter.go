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
	"regexp"
)

type nameFilter struct {
	include *regexp.Regexp
	exclude *regexp.Regexp
}

func newNameFilter(config FilterConfig) (*nameFilter, error) {
	var include, exclude *regexp.Regexp
	var err error

	if config.Include != "" {
		include, err = regexp.Compile(config.Include)
		if err != nil {
			return nil, fmt.Errorf("invalid include filter %q: %w", config.Include, err)
		}
	}
	if config.Exclude != "" {
		exclude, err = regexp.Compile(config.Exclude)
		if err != nil {
			return nil, fmt.Errorf("invalid exclude filter %q: %w", config.Exclude, err)
		}
	}
	if include == nil && exclude == nil {
		return nil, nil
	}
	return &nameFilter{include: include, exclude: exclude}, nil
}

func (f *nameFilter) Match(value string) bool {
	return f.MatchAny(value)
}

func (f *nameFilter) MatchAny(values ...string) bool {
	if f == nil {
		return true
	}
	if f.include != nil {
		matched := false
		for _, value := range values {
			if f.include.MatchString(value) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}
	if f.exclude != nil {
		for _, value := range values {
			if f.exclude.MatchString(value) {
				return false
			}
		}
	}
	return true
}

func dimensionValue(dimension, name string) string {
	return dimension + ":" + name
}

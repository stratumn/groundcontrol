// Copyright 2019 Stratumn
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"errors"
	"regexp"
)

// Errors.
var (
	ErrNoMatch = errors.New("no match")
)

var sourceFileRegexes = []*regexp.Regexp{
	regexp.MustCompile(`\(([^:]+:\d+(:(\d+))?)\)`),
	regexp.MustCompile(`([^\s:()]+:\d+(:(\d+))?)`),
}

// MatchSourceFile tries to find a path to a source file in a string.
// If it finds one, it returns the offsets where the the match begins and ends.
// If there isn't a match, it returns ErrNoMatch.
func MatchSourceFile(str string) (filename string, begin, end int, err error) {
	for _, r := range sourceFileRegexes {
		indexes := r.FindStringSubmatchIndex(str)
		if indexes == nil {
			continue
		}

		begin = indexes[2]
		end = indexes[3]
		filename = str[begin:end]
		return
	}

	err = ErrNoMatch
	return
}

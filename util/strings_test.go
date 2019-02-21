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

import "testing"

func TestMatchSourceFile(t *testing.T) {
	tests := []struct {
		name         string
		str          string
		wantFilename string
		wantBegin    int
		wantEnd      int
		wantErr      bool
	}{{
		"Empty string",
		"",
		"", 0, 0, true,
	}, {
		"Blank string",
		" ",
		"", 0, 0, true,
	}, {
		"Text",
		"hello world",
		"", 0, 0, true,
	}, {
		"Source file",
		"./main.go",
		"./main.go", 0, 9, false,
	}, {
		"Source file with line",
		"./main.go:12",
		"./main.go:12", 0, 12, false,
	}, {
		"Text before source file",
		"\tat ./abc/main.go:12",
		"./abc/main.go:12", 4, 20, false,
	}, {
		"Colon after source file",
		"./main.go:12:",
		"./main.go:12", 0, 12, false,
	}, {
		"Source file with offset",
		"./main.go:12:10",
		"./main.go:12:10", 0, 15, false,
	}, {
		"Text after source file with offset",
		"./main.go:12:10hello",
		"./main.go:12:10", 0, 15, false,
	}, {
		"Source file in parentheses",
		"    at Object.<anonymous> (/src/index.js:12) ",
		"/src/index.js:12", 27, 43, false,
	}, {
		"Source file in parentheses with offset",
		"in (./src/main.go:12:6)",
		"./src/main.go:12:6", 4, 22, false,
	}, {
		"Text in parentheses",
		"in (test)",
		"", 0, 0, true,
	}, {
		"Two source files",
		"./main.go /util.go",
		"./main.go", 0, 9, false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFilename, gotBegin, gotEnd, err := MatchSourceFile(tt.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("MatchSourceFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFilename != tt.wantFilename {
				t.Errorf("MatchSourceFile() gotFilename = %v, want %v", gotFilename, tt.wantFilename)
			}
			if gotBegin != tt.wantBegin {
				t.Errorf("MatchSourceFile() gotBegin = %v, want %v", gotBegin, tt.wantBegin)
			}
			if gotEnd != tt.wantEnd {
				t.Errorf("MatchSourceFile() gotEnd = %v, want %v", gotEnd, tt.wantEnd)
			}
		})
	}
}

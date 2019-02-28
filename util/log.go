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
	"bufio"
	"context"
	"io"
	"time"
)

// LineWriter is used by LineSplitter to output a single line of text to a Logger.
type LineWriter func(ctx context.Context, ownerID, msg string, a ...interface{}) string

// LineSplitter creates a writer with a line splitter that can be used to output lines
// one at a time to a Logger. Remember to call close().
func LineSplitter(ctx context.Context, write LineWriter, ownerID string) io.WriteCloser {
	r, w := io.Pipe()
	scanner := bufio.NewScanner(r)
	go func() {
		for scanner.Scan() {
			write(ctx, ownerID, scanner.Text())
			// Don't kill the poor browser.
			time.Sleep(10 * time.Millisecond)
		}
	}()
	return w
}

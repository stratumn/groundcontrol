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

package pubsub

import (
	"sync"
)

type record struct {
	id      uint64
	message interface{}
}

type history struct {
	cap int

	mu      sync.RWMutex
	records []record
	head    int
}

func newHistory(cap int) *history {
	return &history{
		cap: cap,
	}
}

func (h *history) Add(id uint64, message interface{}) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.records == nil {
		h.records = make([]record, h.cap*2)
	}
	if h.head >= h.cap*2 {
		copy(h.records, h.records[h.cap:])
		h.head = h.cap
	}
	h.records[h.head] = record{
		id:      id,
		message: message,
	}
	h.head++
}

func (h *history) Since(id uint64) (messages []interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	start := h.head - 1
	for ; start >= 0; start-- {
		record := h.records[start]
		if record.id <= id {
			break
		}
	}
	start++
	for ; start < h.head; start++ {
		record := h.records[start]
		messages = append(messages, record.message)
	}
	return
}

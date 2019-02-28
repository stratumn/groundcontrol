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

package store

import (
	"sync"
)

// Memory stores nodes indexed by their ID using an in-memory map.
type Memory struct {
	store sync.Map
	locks sync.Map
}

// NewMemory creates a Memory.
func NewMemory() *Memory {
	return &Memory{}
}

// Store stores a node.
func (s *Memory) Store(id string, node Node) {
	s.store.Store(id, node)
}

// Load loads a node.
func (s *Memory) Load(id string) (Node, bool) {
	node, ok := s.store.Load(id)
	if ok {
		return node.(Node), true
	}

	return nil, false
}

// MustLoad loads a node or panics if it doesn't exist.
func (s *Memory) MustLoad(id string) Node {
	node, ok := s.Load(id)
	if !ok {
		panic(ErrNotFound)
	}

	return node
}

// Delete deletes a node.
func (s *Memory) Delete(id string) {
	s.store.Delete(id)
}

// Lock locks the given IDs.
func (s *Memory) Lock(ids ...string) {
	for _, id := range ids {
		actual, _ := s.locks.LoadOrStore(id, &sync.Mutex{})
		actual.(*sync.Mutex).Lock()
	}
}

// Unlock unlocks the given IDs.
func (s *Memory) Unlock(ids ...string) {
	for _, id := range ids {
		actual, ok := s.locks.Load(id)
		if !ok {
			panic("attempted to unlock unlocked ID")
		}

		actual.(*sync.Mutex).Unlock()
	}
}

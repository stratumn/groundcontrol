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

package models

import (
	"sync"
)

// NodeManager helps manage nodes with global IDs.
type NodeManager struct {
	store sync.Map
	locks sync.Map
}

// NewNodeManager creates a NodeManager.
func NewNodeManager() *NodeManager {
	return &NodeManager{}
}

// Store stores a node.
func (n *NodeManager) Store(id string, node Node) {
	n.store.Store(id, node)
}

// Load loads a node.
func (n *NodeManager) Load(id string) (Node, bool) {
	node, ok := n.store.Load(id)
	if ok {
		return node.(Node), true
	}

	return nil, false
}

// MustLoad loads a node or panics if it doesn't exist.
func (n *NodeManager) MustLoad(id string) Node {
	node, ok := n.Load(id)
	if !ok {
		panic(ErrNotFound)
	}

	return node
}

// Delete deletes a node.
func (n *NodeManager) Delete(id string) {
	n.store.Delete(id)
}

// Lock locks the given IDs.
func (n *NodeManager) Lock(ids ...string) {
	for _, id := range ids {
		actual, _ := n.locks.LoadOrStore(id, &sync.Mutex{})
		actual.(*sync.Mutex).Lock()
	}
}

// Unlock unlocks the given IDs.
func (n *NodeManager) Unlock(ids ...string) {
	for _, id := range ids {
		actual, ok := n.locks.Load(id)
		if !ok {
			panic("attempted to unlock unlocked ID")
		}

		actual.(*sync.Mutex).Unlock()
	}
}

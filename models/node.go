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
	"errors"
	"sync"

	"github.com/stratumn/groundcontrol/relay"
)

// Types.
var (
	UserType       = "User"
	WorkspaceType  = "Workspace"
	ProjectType    = "Project"
	CommitType     = "Commit"
	JobType        = "Job"
	JobMetricsType = "JobMetrics"
)

// Errors.
var (
	ErrNotFound = errors.New("not found")
	ErrType     = errors.New("wrong type")
)

// NodeManager helps manage nodes with global IDs.
type NodeManager struct {
	store sync.Map
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

// LoadUser loads a user.
func (n *NodeManager) LoadUser(id string) (*User, error) {
	identifiers, err := relay.DecodeID(id)
	if err != nil {
		return nil, err
	}
	if identifiers[0] != UserType {
		return nil, ErrType
	}
	node, ok := n.store.Load(id)
	if !ok {
		return nil, ErrNotFound
	}

	return node.(*User), nil
}

// LoadWorkspace loads a workspace.
func (n *NodeManager) LoadWorkspace(id string) (*Workspace, error) {
	identifiers, err := relay.DecodeID(id)
	if err != nil {
		return nil, err
	}
	if identifiers[0] != WorkspaceType {
		return nil, ErrType
	}
	node, ok := n.store.Load(id)
	if !ok {
		return nil, ErrNotFound
	}

	return node.(*Workspace), nil
}

// LoadProject loads a project.
func (n *NodeManager) LoadProject(id string) (*Project, error) {
	identifiers, err := relay.DecodeID(id)
	if err != nil {
		return nil, err
	}
	if identifiers[0] != ProjectType {
		return nil, ErrType
	}
	node, ok := n.store.Load(id)
	if !ok {
		return nil, ErrNotFound
	}

	return node.(*Project), nil
}

// LoadCommit loads a commit.
func (n *NodeManager) LoadCommit(id string) (*Commit, error) {
	identifiers, err := relay.DecodeID(id)
	if err != nil {
		return nil, err
	}
	if identifiers[0] != CommitType {
		return nil, ErrType
	}
	node, ok := n.store.Load(id)
	if !ok {
		return nil, ErrNotFound
	}

	return node.(*Commit), nil
}

// LoadJob loads a job.
func (n *NodeManager) LoadJob(id string) (*Job, error) {
	identifiers, err := relay.DecodeID(id)
	if err != nil {
		return nil, err
	}
	if identifiers[0] != JobType {
		return nil, ErrType
	}
	node, ok := n.store.Load(id)
	if !ok {
		return nil, ErrNotFound
	}

	return node.(*Job), nil
}

// LoadJobMetrics loads a job metrics.
func (n *NodeManager) LoadJobMetrics(id string) (*JobMetrics, error) {
	identifiers, err := relay.DecodeID(id)
	if err != nil {
		return nil, err
	}
	if identifiers[0] != JobMetricsType {
		return nil, ErrType
	}
	node, ok := n.store.Load(id)
	if !ok {
		return nil, ErrNotFound
	}

	return node.(*JobMetrics), nil
}

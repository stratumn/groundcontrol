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
	"context"

	"github.com/stratumn/groundcontrol/relay"
)

// Source is a collection of workspaces.
type Source interface {
	Node
	IsSource()

	// GetWorkspaceIDs returns the IDs of the workspaces.
	GetWorkspaceIDs() []string

	// Workspaces are the workspaces using Relay pagination.
	Workspaces(
		ctx context.Context,
		after *string,
		before *string,
		first *int,
		last *int,
	) (WorkspaceConnection, error)
}

// MustLoadSource loads a Source or panics on failure.
func (n *NodeManager) MustLoadSource(id string) Source {
	identifiers, err := relay.DecodeID(id)
	if err != nil {
		panic(err)
	}

	switch identifiers[0] {
	case NodeTypeDirectorySource:
		return n.MustLoadDirectorySource(id)
	case NodeTypeGitSource:
		return n.MustLoadGitSource(id)
	}

	panic(ErrType)
}

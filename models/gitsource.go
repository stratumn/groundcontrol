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

import "context"

// GitSource is a collection of workspaces in a Git repository.
type GitSource struct {
	// The global ID of the node.
	ID string `json:"id"`
	// The IDs of the workspaces.
	WorkspaceIDs []string `json:"workspaceIds"`
	// Whether currently loading workspaces.
	IsLoading bool `json:"isLoading"`
	// The Git repository.
	Repository string `json:"repository"`
	// The Git branch.
	Branch string `json:"branch"`
}

// IsNode tells gqlgen that it implements Node.
func (GitSource) IsNode() {}

// IsSource tells gqlgen that it implements Source.
func (GitSource) IsSource() {}

// GetWorkspaceIDs returns the IDs of the workspaces.
func (n GitSource) GetWorkspaceIDs() []string {
	return n.WorkspaceIDs
}

// Workspaces are the workspaces using Relay pagination.
func (n GitSource) Workspaces(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (WorkspaceConnection, error) {
	return PaginateWorkspaceIDSliceContext(
		ctx,
		n.WorkspaceIDs,
		after,
		before,
		first,
		last,
	)
}

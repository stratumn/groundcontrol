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

// DirectorySource is a collection of workspaces in a directory.
type DirectorySource struct {
	// The global ID of the node.
	ID string `json:"id"`
	// The ID of the user.
	UserID string `json:"userId"`
	// The IDs of the workspaces.
	WorkspaceIDs []string `json:"workspaceIds"`
	// Whether currently loading workspaces.
	IsLoading bool `json:"isLoading"`
	// The path to the directory containing the workspaces.
	Directory string `json:"directory"`
}

// IsNode tells gqlgen that it implements Node.
func (DirectorySource) IsNode() {}

// IsSource tells gqlgen that it implements Source.
func (DirectorySource) IsSource() {}

// User returns the user who owns the source.
func (n DirectorySource) User(ctx context.Context) User {
	return MustLoadUser(ctx, n.UserID)
}

// GetWorkspaceIDs returns the IDs of the workspaces.
func (n DirectorySource) GetWorkspaceIDs() []string {
	return n.WorkspaceIDs
}

// Workspaces are the workspaces using Relay pagination.
func (n DirectorySource) Workspaces(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (WorkspaceConnection, error) {
	return PaginateWorkspaceIDSlice(
		ctx,
		n.WorkspaceIDs,
		after,
		before,
		first,
		last,
	)
}

// Update loads the workspaces and upserts the source.
func (n *DirectorySource) Update(ctx context.Context) error {
	n.IsLoading = true
	n.MustStore(ctx)

	defer func() {
		n.IsLoading = false
		n.MustStore(ctx)
	}()

	workspaceIDs, err := LoadWorkspacesInSource(ctx, n.Directory, n.ID)
	if err != nil {
		return err
	}

	n.WorkspaceIDs = workspaceIDs
	n.MustStore(ctx)

	return nil
}

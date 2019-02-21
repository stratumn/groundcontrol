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

	"groundcontrol/util"

	"gopkg.in/src-d/go-git.v4/plumbing"
)

// GitSource is a collection of workspaces in a Git repository.
type GitSource struct {
	// The global ID of the node.
	ID string `json:"id"`
	// The ID of the user.
	UserID string `json:"userId"`
	// The IDs of the workspaces.
	WorkspaceIDs []string `json:"workspaceIds"`
	// Whether currently loading workspaces.
	IsLoading bool `json:"isLoading"`
	// The Git repository.
	Repository string `json:"repository"`
	// The Git reference.
	Reference string `json:"reference"`
}

// IsNode tells gqlgen that it implements Node.
func (GitSource) IsNode() {}

// IsSource tells gqlgen that it implements Source.
func (GitSource) IsSource() {}

// User returns the user who owns the source.
func (n GitSource) User(ctx context.Context) User {
	return MustLoadUser(ctx, n.UserID)
}

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
	return PaginateWorkspaceIDSlice(
		ctx,
		n.WorkspaceIDs,
		after,
		before,
		first,
		last,
	)
}

// IsCloned checks if the project is cloned.
func (n GitSource) IsCloned(ctx context.Context) bool {
	getGitSourcePath := GetModelContext(ctx).GetGitSourcePath
	directory := getGitSourcePath(n.Repository, n.Reference)

	return util.FileExists(directory)
}

// ReferenceShort returns the short name of the reference.
func (n GitSource) ReferenceShort() string {
	return plumbing.ReferenceName(n.Reference).Short()
}

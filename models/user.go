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

// Workspaces lists the workspaces belonging to the User using Relay pagination.
func (n *User) Workspaces(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (*WorkspaceConnection, error) {
	return PaginateWorkspaceIDSlice(ctx, n.WorkspacesIDs(ctx), after, before, first, last, nil)
}

// WorkspacesIDs returns the IDs of the workspaces belonging to the User.
func (n *User) WorkspacesIDs(ctx context.Context) []string {
	var slice []string

	for _, sourceID := range n.SourcesIDs {
		source := MustLoadSource(ctx, sourceID)
		slice = append(slice, source.GetWorkspacesIDs()...)
	}

	return slice
}

// Workspace find a workspace by its slug.
func (n *User) Workspace(ctx context.Context, slug string) *Workspace {
	for _, id := range n.WorkspacesIDs(ctx) {
		node := MustLoadWorkspace(ctx, id)

		if node.Slug == slug {
			return node
		}
	}

	return nil
}

// Projects lists the projects belonging to the User using Relay pagination.
func (n *User) Projects(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (*ProjectConnection, error) {
	var slice []string

	for _, workspaceID := range n.WorkspacesIDs(ctx) {
		workspace := MustLoadWorkspace(ctx, workspaceID)
		slice = append(slice, workspace.ProjectsIDs...)
	}

	return PaginateProjectIDSlice(ctx, slice, after, before, first, last, nil)
}

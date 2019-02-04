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

// User contains all the data of the person using the app.
type User struct {
	ID           string   `json:"id"`
	SourceIDs    []string `json:"sourceIds"`
	WorkspaceIDs []string `json:"workspaceIds"`
}

// IsNode tells gqlgen that it implements Node.
func (User) IsNode() {}

// Sources returns the user's sources.
func (u User) Sources(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (SourceConnection, error) {
	return PaginateSourceIDSliceContext(ctx, u.SourceIDs, after, before, first, last)
}

// Workspaces returns the user's workspaces.
func (u User) Workspaces(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (WorkspaceConnection, error) {
	return PaginateWorkspaceIDSliceContext(ctx, u.WorkspaceIDs, after, before, first, last)
}

// Workspace finds a workspace.
func (u User) Workspace(ctx context.Context, slug string) *Workspace {
	nodes := GetModelContext(ctx).Nodes

	for _, id := range u.WorkspaceIDs {
		node := nodes.MustLoadWorkspace(id)

		if node.Slug == slug {
			return &node
		}
	}

	return nil
}

// Projects returns the user's projects.
func (u User) Projects(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (ProjectConnection, error) {
	var slice []string

	nodes := GetModelContext(ctx).Nodes

	for _, workspaceID := range u.WorkspaceIDs {
		workspace := nodes.MustLoadWorkspace(workspaceID)
		slice = append(slice, workspace.ProjectIDs...)
	}

	return PaginateProjectIDSlice(nodes, slice, after, before, first, last)
}

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
	ID        string   `json:"id"`
	SourceIDs []string `json:"sourceIds"`
	KeyIDs    []string `json:"keyIds"`
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
	return PaginateSourceIDSlice(ctx, u.SourceIDs, after, before, first, last)
}

// WorkspaceIDs returns the user's workspace IDs.
func (u User) WorkspaceIDs(ctx context.Context) []string {
	var slice []string

	for _, sourceID := range u.SourceIDs {
		source := MustLoadSource(ctx, sourceID)
		slice = append(slice, source.GetWorkspaceIDs()...)
	}

	return slice
}

// Workspaces returns the user's workspaces.
func (u User) Workspaces(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (WorkspaceConnection, error) {
	return PaginateWorkspaceIDSlice(
		ctx,
		u.WorkspaceIDs(ctx),
		after,
		before,
		first,
		last,
	)
}

// Workspace finds a workspace.
func (u User) Workspace(ctx context.Context, slug string) *Workspace {
	for _, id := range u.WorkspaceIDs(ctx) {
		node := MustLoadWorkspace(ctx, id)

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

	for _, workspaceID := range u.WorkspaceIDs(ctx) {
		workspace := MustLoadWorkspace(ctx, workspaceID)
		slice = append(slice, workspace.ProjectIDs...)
	}

	return PaginateProjectIDSlice(ctx, slice, after, before, first, last)
}

// Keys returns the user's keys.
func (u User) Keys(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (KeyConnection, error) {
	return PaginateKeyIDSlice(
		ctx,
		u.KeyIDs,
		after,
		before,
		first,
		last,
	)
}

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

package model

import (
	"context"
	"sort"
	"strings"
)

// BeforeStore sorts collections before storing the user.
func (n *User) BeforeStore(ctx context.Context) {
	sort.Slice(n.SourcesIDs, func(i, j int) bool {
		a := MustLoadSource(ctx, n.SourcesIDs[i])
		b := MustLoadSource(ctx, n.SourcesIDs[j])
		u, v := "", ""
		switch source := a.(type) {
		case *DirectorySource:
			u = source.Directory
		case *GitSource:
			u = source.Repository
		}
		switch source := b.(type) {
		case *DirectorySource:
			v = source.Directory
		case *GitSource:
			v = source.Repository
		}
		return strings.ToLower(u) < strings.ToLower(v)
	})

	sort.Slice(n.KeysIDs, func(i, j int) bool {
		a := MustLoadKey(ctx, n.KeysIDs[i])
		b := MustLoadKey(ctx, n.KeysIDs[j])
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})
}

// Workspaces lists the Workspaces belonging to the User sorted by Name
// using Relay pagination.
func (n *User) Workspaces(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (*WorkspaceConnection, error) {
	return PaginateWorkspaceIDSlice(ctx, n.WorkspacesIDs(ctx), after, before, first, last, nil)
}

// WorkspacesIDs returns the IDs of the Workspaces belonging to the User
// sorted by Name.
func (n *User) WorkspacesIDs(ctx context.Context) []string {
	var slice []string

	for _, sourceID := range n.SourcesIDs {
		source := MustLoadSource(ctx, sourceID)
		slice = append(slice, source.GetWorkspacesIDs()...)
	}

	sort.Slice(slice, func(i, j int) bool {
		a := MustLoadWorkspace(ctx, slice[i])
		b := MustLoadWorkspace(ctx, slice[j])
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})

	return slice
}

// Workspace find a Workspace by its slug.
func (n *User) Workspace(ctx context.Context, slug string) *Workspace {
	for _, id := range n.WorkspacesIDs(ctx) {
		node := MustLoadWorkspace(ctx, id)

		if node.Slug == slug {
			return node
		}
	}

	return nil
}

// Projects lists the Projects belonging to the User using Relay pagination.
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

// Services lists the Services belonging to the User using Relay pagination
// optionally filtered by ServiceStatus.
func (n *User) Services(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
	status []ServiceStatus,
) (*ServiceConnection, error) {
	var slice []string

	for _, workspaceID := range n.WorkspacesIDs(ctx) {
		workspace := MustLoadWorkspace(ctx, workspaceID)
		slice = append(slice, workspace.ServicesIDs...)
	}

	filter := func(node *Service) bool {
		return n.filterServiceNode(ctx, node, status)
	}

	return PaginateServiceIDSlice(ctx, slice, after, before, first, last, filter)
}

func (n *User) filterServiceNode(ctx context.Context, node *Service, status []ServiceStatus) bool {
	match := len(status) == 0

	for _, v := range status {
		if node.Status == v {
			match = true
			break
		}
	}

	return match
}

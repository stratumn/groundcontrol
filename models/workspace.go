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

// Workspace represents a workspace in the app.
type Workspace struct {
	ID          string   `json:"id"`
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	ProjectIDs  []string `json:"projectIDs"`
	TaskIDs     []string `json:"taskIDs"`
	Description string   `json:"description"`
	Notes       *string  `json:"notes"`
}

// IsNode tells gqlgen that it implements Node.
func (Workspace) IsNode() {}

// Projects returns the workspace's projects.
func (w Workspace) Projects(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (ProjectConnection, error) {
	return PaginateProjectIDSliceContext(ctx, w.ProjectIDs, after, before, first, last)
}

// Tasks returns the workspace's tasks.
func (w Workspace) Tasks(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (TaskConnection, error) {
	return PaginateTaskIDSliceContext(ctx, w.TaskIDs, after, before, first, last)
}

// IsCloning returns true if any of the projects is cloning.
func (w Workspace) IsCloning(ctx context.Context) bool {
	nodes := GetModelContext(ctx).Nodes

	for _, id := range w.ProjectIDs {
		node := nodes.MustLoadProject(id)
		if node.IsCloning {
			return true
		}
	}

	return false
}

// IsCloned returns true if all the projects are cloned.
func (w Workspace) IsCloned(ctx context.Context) bool {
	nodes := GetModelContext(ctx).Nodes

	for _, id := range w.ProjectIDs {
		node := nodes.MustLoadProject(id)
		if !node.IsCloned(ctx) {
			return false
		}
	}

	return true
}

// IsPulling returns true if any of the projects is pulling.
func (w Workspace) IsPulling(ctx context.Context) bool {
	nodes := GetModelContext(ctx).Nodes

	for _, id := range w.ProjectIDs {
		node := nodes.MustLoadProject(id)
		if node.IsPulling {
			return true
		}
	}

	return false
}

// IsBehind returns true if any of the projects is behind origin.
func (w Workspace) IsBehind(ctx context.Context) bool {
	nodes := GetModelContext(ctx).Nodes

	for _, id := range w.ProjectIDs {
		node := nodes.MustLoadProject(id)
		if node.IsBehind {
			return true
		}
	}

	return false
}

// IsAhead returns true if any of the projects is ahead origin.
func (w Workspace) IsAhead(ctx context.Context) bool {
	nodes := GetModelContext(ctx).Nodes

	for _, id := range w.ProjectIDs {
		node := nodes.MustLoadProject(id)
		if node.IsAhead {
			return true
		}
	}

	return false
}

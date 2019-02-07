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

// Task represents a workspace task in the app.
type Task struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	VariableIDs []string `json:"variableIds"`
	StepIDs     []string `json:"stepIds"`
	WorkspaceID string   `json:"workspace"`
	IsRunning   bool     `json:"isRunning"`
}

// IsNode tells gqlgen that it implements Node.
func (Task) IsNode() {}

// Variables returns the task's variables.
func (t Task) Variables(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (VariableConnection, error) {
	return PaginateVariableIDSliceContext(ctx, t.VariableIDs, after, before, first, last)
}

// Steps returns the task's steps.
func (t Task) Steps(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (StepConnection, error) {
	return PaginateStepIDSliceContext(ctx, t.StepIDs, after, before, first, last)
}

// Workspace returns the task's workspace.
func (t Task) Workspace(ctx context.Context) Workspace {
	return GetModelContext(ctx).Nodes.MustLoadWorkspace(t.WorkspaceID)
}

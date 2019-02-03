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

// Step represents a task step in the app.
type Step struct {
	ID         string   `json:"id"`
	ProjectIDs []string `json:"projectIds"`
	CommandIDs []string `json:"commandIds"`
	TaskID     string   `json:"taskId"`
}

// IsNode tells gqlgen that it implements Node.
func (Step) IsNode() {}

// Projects returns the step's projects.
func (s Step) Projects(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (ProjectConnection, error) {
	return PaginateProjectIDSliceContext(ctx, s.ProjectIDs, after, before, first, last)
}

// Commands returns the step's commands.
func (s Step) Commands(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (CommandConnection, error) {
	return PaginateCommandIDSliceContext(ctx, s.ProjectIDs, after, before, first, last)
}

// Task returns the step's taks.
func (s Step) Task(ctx context.Context) Task {
	return GetModelContext(ctx).Nodes.MustLoadTask(s.TaskID)
}

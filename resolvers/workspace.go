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

package resolvers

import (
	"context"

	"github.com/stratumn/groundcontrol/models"
)

type workspaceResolver struct {
	*Resolver
}

func (r *workspaceResolver) Projects(ctx context.Context, obj *models.Workspace, after, before *string, first, last *int) (models.ProjectConnection, error) {
	return obj.Projects(r.Nodes, after, before, first, last)
}

func (r *workspaceResolver) Tasks(ctx context.Context, obj *models.Workspace, after, before *string, first, last *int) (models.TaskConnection, error) {
	return obj.Tasks(r.Nodes, after, before, first, last)
}

func (r *workspaceResolver) IsCloning(ctx context.Context, obj *models.Workspace) (bool, error) {
	return obj.IsCloning(r.Nodes), nil
}

func (r *workspaceResolver) IsCloned(ctx context.Context, obj *models.Workspace) (bool, error) {
	return obj.IsCloned(r.Nodes, r.GetProjectPath), nil
}

func (r *workspaceResolver) IsPulling(ctx context.Context, obj *models.Workspace) (bool, error) {
	return obj.IsPulling(r.Nodes), nil
}

func (r *workspaceResolver) IsBehind(ctx context.Context, obj *models.Workspace) (bool, error) {
	return obj.IsBehind(r.Nodes), nil
}

func (r *workspaceResolver) IsAhead(ctx context.Context, obj *models.Workspace) (bool, error) {
	return obj.IsAhead(r.Nodes), nil
}

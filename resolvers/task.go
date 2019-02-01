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

type taskResolver struct {
	*Resolver
}

func (r *taskResolver) Steps(ctx context.Context, obj *models.Task, after, before *string, first, last *int) (models.StepConnection, error) {
	return obj.Steps(r.Nodes, after, before, first, last)
}

func (r *taskResolver) Workspace(ctx context.Context, obj *models.Task) (models.Workspace, error) {
	return obj.Workspace(r.Nodes), nil
}

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

type processGroupResolver struct {
	*Resolver
}

func (r *processGroupResolver) Processes(ctx context.Context, obj *models.ProcessGroup, after, before *string, first, last *int) (models.ProcessConnection, error) {
	return obj.Processes(r.Nodes, after, before, first, last)
}

func (r *processGroupResolver) Task(ctx context.Context, obj *models.ProcessGroup) (models.Task, error) {
	return obj.Task(r.Nodes), nil
}

func (r *processGroupResolver) Status(ctx context.Context, obj *models.ProcessGroup) (models.ProcessStatus, error) {
	return obj.Status(r.Nodes), nil
}

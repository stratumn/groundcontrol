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

type subscriptionResolver struct{ *Resolver }

func (r *subscriptionResolver) WorkspaceUpdated(
	ctx context.Context,
	id *string,
) (<-chan models.Workspace, error) {
	ch := make(chan models.Workspace)

	r.PubSub.Subscribe(ctx, models.WorkspaceUpdated, func(msg interface{}) {
		workspace := msg.(*models.Workspace)
		if id != nil && *id != workspace.ID {
			return
		}
		select {
		case <-ctx.Done():
		case ch <- *workspace:
		}
	})

	return ch, nil
}

func (r *subscriptionResolver) ProjectUpdated(
	ctx context.Context,
	id *string,
) (<-chan models.Project, error) {
	ch := make(chan models.Project)

	r.PubSub.Subscribe(ctx, models.ProjectUpdated, func(msg interface{}) {
		project := msg.(*models.Project)
		if id != nil && *id != project.ID {
			return
		}
		select {
		case <-ctx.Done():
		case ch <- *project:
		}
	})

	return ch, nil
}

func (r *subscriptionResolver) JobUpserted(
	ctx context.Context,
) (<-chan models.Job, error) {
	ch := make(chan models.Job)

	r.PubSub.Subscribe(ctx, models.JobUpserted, func(msg interface{}) {
		select {
		case <-ctx.Done():
		case ch <- *msg.(*models.Job):
		}
	})

	return ch, nil
}

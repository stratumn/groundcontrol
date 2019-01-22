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

type subscriptionResolver struct {
	*Resolver
}

func (r *subscriptionResolver) WorkspaceUpdated(
	ctx context.Context,
	id *string,
) (<-chan models.Workspace, error) {
	ch := make(chan models.Workspace)

	r.Subs.Subscribe(ctx, models.WorkspaceUpdated, func(msg interface{}) {
		nodeID := msg.(string)
		if id != nil && *id != nodeID {
			return
		}
		select {
		case <-ctx.Done():
		case ch <- r.Nodes.MustLoadWorkspace(nodeID):
		}
	})

	return ch, nil
}

func (r *subscriptionResolver) ProjectUpdated(
	ctx context.Context,
	id *string,
) (<-chan models.Project, error) {
	ch := make(chan models.Project)

	r.Subs.Subscribe(ctx, models.ProjectUpdated, func(msg interface{}) {
		nodeID := msg.(string)
		if id != nil && *id != nodeID {
			return
		}
		select {
		case <-ctx.Done():
		case ch <- r.Nodes.MustLoadProject(nodeID):
		}
	})

	return ch, nil
}

func (r *subscriptionResolver) JobUpserted(
	ctx context.Context,
) (<-chan models.Job, error) {
	ch := make(chan models.Job)

	r.Subs.Subscribe(ctx, models.JobUpserted, func(msg interface{}) {
		select {
		case <-ctx.Done():
		case ch <- r.Nodes.MustLoadJob(msg.(string)):
		}
	})

	return ch, nil
}

func (r *subscriptionResolver) JobMetricsUpdated(
	ctx context.Context,
) (<-chan models.JobMetrics, error) {
	ch := make(chan models.JobMetrics)

	r.Subs.Subscribe(ctx, models.JobMetricsUpdated, func(msg interface{}) {
		select {
		case <-ctx.Done():
		case ch <- r.Nodes.MustLoadJobMetrics(msg.(string)):
		}
	})

	return ch, nil
}

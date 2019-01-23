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

// SubscriptionChannelSize is the size of a subscription channel.
// If a channel is full new messages will be dropped and the client won't receive them.
const SubscriptionChannelSize = 128

type subscriptionResolver struct {
	*Resolver
}

func (r *subscriptionResolver) WorkspaceUpdated(
	ctx context.Context,
	id *string,
) (<-chan models.Workspace, error) {
	ch := make(chan models.Workspace, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.WorkspaceUpdated, func(msg interface{}) {
		nodeID := msg.(string)
		if id != nil && *id != nodeID {
			return
		}
		select {
		case ch <- r.Nodes.MustLoadWorkspace(nodeID):
		default:
		}
	})

	return ch, nil
}

func (r *subscriptionResolver) ProjectUpdated(
	ctx context.Context,
	id *string,
) (<-chan models.Project, error) {
	ch := make(chan models.Project, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.ProjectUpdated, func(msg interface{}) {
		nodeID := msg.(string)
		if id != nil && *id != nodeID {
			return
		}
		select {
		case ch <- r.Nodes.MustLoadProject(nodeID):
		default:
		}
	})

	return ch, nil
}

func (r *subscriptionResolver) JobUpserted(
	ctx context.Context,
) (<-chan models.Job, error) {
	ch := make(chan models.Job, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.JobUpserted, func(msg interface{}) {
		select {
		case ch <- r.Nodes.MustLoadJob(msg.(string)):
		default:
		}
	})

	return ch, nil
}

func (r *subscriptionResolver) JobMetricsUpdated(
	ctx context.Context,
) (<-chan models.JobMetrics, error) {
	ch := make(chan models.JobMetrics, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.JobMetricsUpdated, func(msg interface{}) {
		select {
		case ch <- r.Nodes.MustLoadJobMetrics(msg.(string)):
		default:
		}
	})

	return ch, nil
}

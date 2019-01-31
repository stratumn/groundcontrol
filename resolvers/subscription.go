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
const SubscriptionChannelSize = 1024

type subscriptionResolver struct {
	*Resolver
}

func (r *subscriptionResolver) WorkspaceUpdated(
	ctx context.Context,
	id *string,
) (<-chan models.Workspace, error) {
	go func() {
		<-ctx.Done()
	}()

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
	go func() {
		<-ctx.Done()
	}()

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

func (r *subscriptionResolver) TaskUpdated(
	ctx context.Context,
	id *string,
) (<-chan models.Task, error) {
	go func() {
		<-ctx.Done()
	}()

	ch := make(chan models.Task, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.TaskUpdated, func(msg interface{}) {
		nodeID := msg.(string)
		if id != nil && *id != nodeID {
			return
		}
		select {
		case ch <- r.Nodes.MustLoadTask(nodeID):
		default:
		}
	})

	return ch, nil
}

func (r *subscriptionResolver) JobUpserted(
	ctx context.Context,
	id *string,
) (<-chan models.Job, error) {
	go func() {
		<-ctx.Done()
	}()

	ch := make(chan models.Job, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.JobUpserted, func(msg interface{}) {
		nodeID := msg.(string)
		if id != nil && *id != nodeID {
			return
		}
		select {
		case ch <- r.Nodes.MustLoadJob(nodeID):
		default:
		}
	})

	return ch, nil
}

func (r *subscriptionResolver) JobMetricsUpdated(
	ctx context.Context,
	id *string,
) (<-chan models.JobMetrics, error) {
	go func() {
		<-ctx.Done()
	}()

	ch := make(chan models.JobMetrics, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.JobMetricsUpdated, func(msg interface{}) {
		nodeID := msg.(string)
		if id != nil && *id != nodeID {
			return
		}
		select {
		case ch <- r.Nodes.MustLoadJobMetrics(nodeID):
		default:
		}
	})

	return ch, nil
}

func (r *subscriptionResolver) ProcessGroupUpserted(
	ctx context.Context,
	id *string,
) (<-chan models.ProcessGroup, error) {
	go func() {
		<-ctx.Done()
	}()

	ch := make(chan models.ProcessGroup, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.ProcessGroupUpserted, func(msg interface{}) {
		nodeID := msg.(string)
		if id != nil && *id != nodeID {
			return
		}
		select {
		case ch <- r.Nodes.MustLoadProcessGroup(nodeID):
		default:
		}
	})

	return ch, nil
}

func (r *subscriptionResolver) ProcessUpserted(
	ctx context.Context,
	id *string,
) (<-chan models.Process, error) {
	go func() {
		<-ctx.Done()
	}()

	ch := make(chan models.Process, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.ProcessUpserted, func(msg interface{}) {
		nodeID := msg.(string)
		if id != nil && *id != nodeID {
			return
		}
		select {
		case ch <- r.Nodes.MustLoadProcess(nodeID):
		default:
		}
	})

	return ch, nil
}

func (r *subscriptionResolver) ProcessMetricsUpdated(
	ctx context.Context,
	id *string,
) (<-chan models.ProcessMetrics, error) {
	go func() {
		<-ctx.Done()
	}()

	ch := make(chan models.ProcessMetrics, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.ProcessMetricsUpdated, func(msg interface{}) {
		nodeID := msg.(string)
		if id != nil && *id != nodeID {
			return
		}
		select {
		case ch <- r.Nodes.MustLoadProcessMetrics(nodeID):
		default:
		}
	})

	return ch, nil
}

// Note: don't log log events, it would go in infinite loop.

func (r *subscriptionResolver) LogEntryAdded(
	ctx context.Context,
) (<-chan models.LogEntry, error) {
	ch := make(chan models.LogEntry, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.LogEntryAdded, func(msg interface{}) {
		select {
		case ch <- r.Nodes.MustLoadLogEntry(msg.(string)):
		default:
		}
	})

	return ch, nil
}

func (r *subscriptionResolver) LogMetricsUpdated(
	ctx context.Context,
	id *string,
) (<-chan models.LogMetrics, error) {
	ch := make(chan models.LogMetrics, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.LogMetricsUpdated, func(msg interface{}) {
		nodeID := msg.(string)
		if id != nil && *id != nodeID {
			return
		}
		select {
		case ch <- r.Nodes.MustLoadLogMetrics(nodeID):
		default:
		}
	})

	return ch, nil
}

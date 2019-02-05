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

func (r *subscriptionResolver) SourceDeleted(
	ctx context.Context,
	id *string,
) (<-chan models.DeletedNode, error) {
	ch := make(chan models.DeletedNode, SubscriptionChannelSize)

	r.Subs.Subscribe(ctx, models.SourceDeleted, func(msg interface{}) {
		sourceID := msg.(string)
		if id != nil && *id != sourceID {
			return
		}

		select {
		case ch <- models.DeletedNode{ID: sourceID}:
		default:
		}
	})

	return ch, nil
}

// Define custom subscriptions for the log because we don't want to log them to avoid infinite loops.

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

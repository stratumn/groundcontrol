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

func (r *subscriptionResolver) LogEntryAdded(
	ctx context.Context,
	lastMessageID *string,
) (<-chan models.LogEntry, error) {
	ch := make(chan models.LogEntry, SubscriptionChannelSize)

	last := uint64(0)
	if lastMessageID != nil {
		var err error
		last, err = decodeBase64Uint64(*lastMessageID)
		if err != nil {
			return nil, err
		}
	}

	r.Subs.Subscribe(ctx, models.LogEntryAdded, last, func(msg interface{}) {
		select {
		case ch <- r.Nodes.MustLoadLogEntry(msg.(string)):
		default:
		}
	})

	return ch, nil
}

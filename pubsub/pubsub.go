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

package pubsub

import (
	"context"
	"sync"
	"sync/atomic"
)

// PubSub deals with subscribing and publishing messages.
type PubSub struct {
	historyCap int

	subs   sync.Map
	lastID uint64

	lastMessages  sync.Map
	lastMessageID uint64
}

// New creates a new PubSub.
// It will keep a history for each message type with at least the given cap.
func New(historyCap int) *PubSub {
	return &PubSub{
		historyCap:    historyCap,
		lastMessageID: 1,
	}
}

// Subscribe register a function that will receive messages of the given type.
// To unsubscribe the context must be closed.
func (p *PubSub) Subscribe(
	ctx context.Context,
	messageType string,
	since uint64,
	fn func(interface{}),
) {
	id := atomic.AddUint64(&p.lastID, 1)

	actual, _ := p.subs.LoadOrStore(messageType, &sync.Map{})
	subs := actual.(*sync.Map)
	subs.Store(id, fn)

	if since > 0 {
		actual, _ = p.lastMessages.LoadOrStore(messageType, newHistory(p.historyCap))
		history := actual.(*history)

		for _, message := range history.Since(since) {
			fn(message)
		}
	}

	go func() {
		<-ctx.Done()
		subs.Delete(id)
	}()
}

// Publish will publish a message of the given type to all subscribers for that type.
func (p *PubSub) Publish(messageType string, message interface{}) {
	messageID := atomic.AddUint64(&p.lastMessageID, 1)

	actual, _ := p.lastMessages.LoadOrStore(messageType, newHistory(p.historyCap))
	history := actual.(*history)
	history.Add(messageID, message)

	actual, _ = p.subs.LoadOrStore(messageType, &sync.Map{})
	messageTypeMap := actual.(*sync.Map)

	messageTypeMap.Range(func(_, v interface{}) bool {
		fn := v.(func(interface{}))
		fn(message)
		return true
	})
}

// LastMessageID returns the ID of the last message.
func (p *PubSub) LastMessageID() uint64 {
	return p.lastMessageID
}

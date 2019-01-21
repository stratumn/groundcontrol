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
	subs   sync.Map
	nextID uint64
}

// New creates a new PubSub.
func New() *PubSub {
	return &PubSub{}
}

// Subscribe register a function that will receive messages of the given type.
// To unsubscribe the context must be closed.
func (p *PubSub) Subscribe(ctx context.Context, msgType string, fn func(interface{})) {
	id := atomic.AddUint64(&p.nextID, 1)
	actual, _ := p.subs.LoadOrStore(msgType, &sync.Map{})
	msgTypeMap := actual.(*sync.Map)
	msgTypeMap.Store(id, fn)

	go func() {
		<-ctx.Done()
		msgTypeMap.Delete(id)
	}()
}

// Publish will publish a message of the given type to all subscribers for that type.
func (p *PubSub) Publish(msgType string, msg interface{}) {
	actual, _ := p.subs.LoadOrStore(msgType, &sync.Map{})
	msgTypeMap := actual.(*sync.Map)

	msgTypeMap.Range(func(_, v interface{}) bool {
		fn := v.(func(interface{}))
		fn(msg)
		return true
	})
}

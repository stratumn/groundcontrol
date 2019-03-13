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

package model

import (
	"context"

	"groundcontrol/appcontext"
	"groundcontrol/relay"
)

// NewKey initializes a new Key.
func NewKey(name, value string) *Key {
	return &Key{
		ID:    relay.EncodeID(NodeTypeKey, name),
		Name:  name,
		Value: value,
	}
}

// AfterStore sets the key in the config after being stored.
// It doesn't save the config.
func (n *Key) AfterStore(ctx context.Context) {
	appcontext.Get(ctx).Keys.Set(n.Name, n.Value)
}

// AfterDelete removes the config after being deleted.
// It doesn't save the config.
func (n *Key) AfterDelete(ctx context.Context) {
	appcontext.Get(ctx).Keys.Delete(n.Name)
}

// InjectKeysConfig creates nodes for all the keys in the config.
func InjectKeysConfig(ctx context.Context) {
	var ids []string
	appCtx := appcontext.Get(ctx)
	for n, v := range appCtx.Keys.All() {
		key := NewKey(n, v)
		key.MustStore(ctx)
		ids = append(ids, key.ID)
	}
	MustLockUser(ctx, appCtx.ViewerID, func(node *User) {
		node.KeysIDs = ids
		node.MustStore(ctx)
	})
}

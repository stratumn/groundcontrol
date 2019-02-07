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

package models

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/relay"
)

// KeysConfig contains all the data in a YAML keys config file.
type KeysConfig struct {
	Filename string            `json:"-" yaml:"-"`
	Keys     map[string]string `json:"keys" yaml:"keys"`
}

// UpsertNodes upserts nodes for the content of the keys config.
// The user node must already exists.
func (c *KeysConfig) UpsertNodes(
	nodes *NodeManager,
	subs *pubsub.PubSub,
	userID string,
) error {
	return nodes.MustLockUserE(userID, func(user User) error {
		var keyIDs []string

		for name, value := range c.Keys {
			key := Key{
				ID:    relay.EncodeID(NodeTypeKey, name),
				Name:  name,
				Value: value,
			}

			keyIDs = append(keyIDs, key.ID)
			nodes.MustStoreKey(key)
			subs.Publish(KeyUpserted, key.ID)
		}

		user.KeyIDs = keyIDs
		nodes.MustStoreUser(user)

		return nil
	})
}

// UpsertKey upserts a key.
// It returns the ID of the key.
func (c *KeysConfig) UpsertKey(
	nodes *NodeManager,
	subs *pubsub.PubSub,
	userID string,
	input KeyInput,
) string {
	key := Key{
		ID: relay.EncodeID(
			NodeTypeKey,
			input.Name,
		),
		Name:  input.Name,
		Value: input.Value,
	}

	nodes.MustLockUser(userID, func(user User) {
		c.Keys[input.Name] = input.Value

		for _, keyID := range user.KeyIDs {
			if keyID == key.ID {
				return
			}
		}

		user.KeyIDs = append(user.KeyIDs, key.ID)

		nodes.MustStoreKey(key)
		nodes.MustStoreUser(user)

		subs.Publish(KeyUpserted, key.ID)
	})

	return key.ID
}

// DeleteKey deletes a key.
func (c *KeysConfig) DeleteKey(
	nodes *NodeManager,
	subs *pubsub.PubSub,
	userID string,
	id string,
) error {
	return nodes.LockUserE(userID, func(user User) error {
		return nodes.LockKey(id, func(key Key) {
			for i, v := range user.KeyIDs {
				if v == id {
					user.KeyIDs = append(
						user.KeyIDs[:i],
						user.KeyIDs[i+1:]...,
					)
					break
				}
			}

			delete(c.Keys, key.Name)

			nodes.MustDeleteKey(id)
			nodes.MustStoreUser(user)
			subs.Publish(KeyDeleted, id)
		})
	})
}

// Save saves the config to disk, overwriting the file if it exists.
func (c KeysConfig) Save() error {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(c.Filename), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(c.Filename, bytes, 0600)
}

// LoadKeysConfigYAML loads a key config from a YAML file.
// It will create a file if it doesn't exist.
func LoadKeysConfigYAML(filename string) (*KeysConfig, error) {
	config := KeysConfig{
		Filename: filename,
		Keys:     map[string]string{},
	}

	bytes, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		config := KeysConfig{
			Filename: filename,
		}
		if err := config.Save(); err != nil {
			return nil, err
		}

		return LoadKeysConfigYAML(filename)
	}
	if err != nil {
		return nil, err
	}

	err = yaml.UnmarshalStrict(bytes, &config)

	return &config, err
}

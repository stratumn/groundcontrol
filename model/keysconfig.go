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
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"groundcontrol/relay"
)

// KeysConfig contains all the data in a YAML keys config file.
type KeysConfig struct {
	Filename string            `json:"-" yaml:"-"`
	Keys     map[string]string `json:"keys" yaml:"keys"`
}

// UpsertNodes upserts nodes for the content of the keys config.
func (c *KeysConfig) UpsertNodes(ctx context.Context) error {
	modelCtx := GetContext(ctx)

	return MustLockUserE(ctx, modelCtx.ViewerID, func(viewer *User) error {
		var keysIDs []string

		for name, value := range c.Keys {
			key := Key{
				ID:    relay.EncodeID(NodeTypeKey, name),
				Name:  name,
				Value: value,
			}

			key.MustStore(ctx)
			keysIDs = append(keysIDs, key.ID)
		}

		viewer.KeysIDs = keysIDs
		viewer.MustStore(ctx)

		return nil
	})
}

// UpsertKey upserts a key.
// It returns the ID of the key.
func (c *KeysConfig) UpsertKey(ctx context.Context, input KeyInput) string {
	modelCtx := GetContext(ctx)

	key := Key{
		ID: relay.EncodeID(
			NodeTypeKey,
			input.Name,
		),
		Name:  input.Name,
		Value: input.Value,
	}

	MustLockUser(ctx, modelCtx.ViewerID, func(viewer *User) {
		c.Keys[input.Name] = input.Value

		exists := false
		for _, keysID := range viewer.KeysIDs {
			if keysID == key.ID {
				exists = true
			}
		}

		if !exists {
			viewer.KeysIDs = append(viewer.KeysIDs, key.ID)
		}

		key.MustStore(ctx)
		viewer.MustStore(ctx)
	})

	return key.ID
}

// DeleteKey deletes a key.
func (c *KeysConfig) DeleteKey(ctx context.Context, id string) error {
	modelCtx := GetContext(ctx)

	return LockUserE(ctx, modelCtx.ViewerID, func(viewer *User) error {
		return LockKey(ctx, id, func(key *Key) {
			for i, v := range viewer.KeysIDs {
				if v == id {
					viewer.KeysIDs = append(
						viewer.KeysIDs[:i],
						viewer.KeysIDs[i+1:]...,
					)
					break
				}
			}

			delete(c.Keys, key.Name)

			MustDeleteKey(ctx, id)
			viewer.MustStore(ctx)
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
		return &config, config.Save()
	}
	if err != nil {
		return nil, err
	}

	return &config, yaml.UnmarshalStrict(bytes, &config)
}

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

package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	yaml "gopkg.in/yaml.v2"
)

// Keys stores keys in a YAML file.
type Keys struct {
	mu       sync.RWMutex
	Filename string            `json:"-" yaml:"-"`
	Keys     map[string]string `json:"keys" yaml:"keys"`
}

// Set sets the value of a key.
func (c *Keys) Set(name, value string) {
	c.mu.Lock()
	c.Keys[name] = value
	c.mu.Unlock()
}

// Delete deletes a key.
func (c *Keys) Delete(name string) {
	c.mu.Lock()
	delete(c.Keys, name)
	c.mu.Unlock()
}

// All returns all the keys.
func (c *Keys) All() map[string]string {
	c.mu.RLock()
	keys := map[string]string{}
	for n, v := range c.Keys {
		keys[n] = v
	}
	c.mu.RUnlock()
	return keys
}

// Save saves the keys to disk, overwriting the file if it exists.
func (c *Keys) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.Filename), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(c.Filename, bytes, 0600)
}

// LoadKeysYAML loads a keys from a YAML file.
// It will create a file if it doesn't exist.
func LoadKeysYAML(filename string) (*Keys, error) {
	config := Keys{Filename: filename, Keys: map[string]string{}}
	bytes, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		return &config, config.Save()
	}
	if err != nil {
		return nil, err
	}
	return &config, yaml.UnmarshalStrict(bytes, &config)
}

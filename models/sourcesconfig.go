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

	"groundcontrol/pubsub"
	"groundcontrol/relay"
)

// SourcesConfig contains all the data in a YAML sources config file.
type SourcesConfig struct {
	Filename         string                  `json:"-" yaml:"-"`
	DirectorySources []DirectorySourceConfig `json:"directorySources" yaml:"directory-sources"`
	GitSources       []GitSourceConfig       `json:"gitSources" yaml:"git-sources"`
}

// DirectorySourceConfig contains all the data in a YAML directory source config file.
type DirectorySourceConfig struct {
	Directory string `json:"directory"`
	ID        string `json:"-" yaml:"-"`
}

// GitSourceConfig contains all the data in a YAML Git source config file.
type GitSourceConfig struct {
	Repository string `json:"repository"`
	Reference  string `json:"reference"`
	ID         string `json:"-" yaml:"-"`
}

// UpsertNodes upserts nodes for the content of the sources config.
// The user node must already exists.
func (c *SourcesConfig) UpsertNodes(
	nodes *NodeManager,
	subs *pubsub.PubSub,
	userID string,
) error {
	return nodes.MustLockUserE(userID, func(user User) error {
		var sourceIDs []string

		for i, sourceConfig := range c.DirectorySources {
			source := DirectorySource{
				ID:        relay.EncodeID(NodeTypeDirectorySource, sourceConfig.Directory),
				Directory: sourceConfig.Directory,
			}

			c.DirectorySources[i].ID = source.ID
			sourceIDs = append(sourceIDs, source.ID)
			nodes.MustStoreDirectorySource(source)
			subs.Publish(SourceUpserted, source.ID)
		}

		for i, sourceConfig := range c.GitSources {
			source := GitSource{
				ID: relay.EncodeID(
					NodeTypeGitSource,
					sourceConfig.Repository,
					sourceConfig.Reference,
				),
				Repository: sourceConfig.Repository,
				Reference:  sourceConfig.Reference,
			}

			c.GitSources[i].ID = source.ID
			sourceIDs = append(sourceIDs, source.ID)
			nodes.MustStoreGitSource(source)
			subs.Publish(SourceUpserted, source.ID)
		}

		user.SourceIDs = sourceIDs
		nodes.MustStoreUser(user)

		return nil
	})
}

// UpsertDirectorySource upserts a directory source.
// It returns the ID of the source.
func (c *SourcesConfig) UpsertDirectorySource(
	nodes *NodeManager,
	subs *pubsub.PubSub,
	userID string,
	input DirectorySourceInput,
) string {
	source := DirectorySource{
		ID: relay.EncodeID(
			NodeTypeDirectorySource,
			input.Directory,
		),
		Directory: input.Directory,
	}

	nodes.MustLockUser(userID, func(user User) {
		for _, sourceID := range user.SourceIDs {
			if sourceID == source.ID {
				return
			}
		}

		nodes.MustStoreDirectorySource(source)

		user.SourceIDs = append(user.SourceIDs, source.ID)
		nodes.MustStoreUser(user)

		c.DirectorySources = append(
			c.DirectorySources,
			DirectorySourceConfig{
				Directory: input.Directory,
				ID:        source.ID,
			},
		)

		subs.Publish(SourceUpserted, source.ID)
	})

	return source.ID
}

// UpsertGitSource upserts a repository source.
// It returns the ID of the source.
func (c *SourcesConfig) UpsertGitSource(
	nodes *NodeManager,
	subs *pubsub.PubSub,
	userID string,
	input GitSourceInput,
) string {
	source := GitSource{
		ID: relay.EncodeID(
			NodeTypeGitSource,
			input.Repository,
			input.Reference,
		),
		Repository: input.Repository,
		Reference:  input.Reference,
	}

	nodes.MustLockUser(userID, func(user User) {
		for _, sourceID := range user.SourceIDs {
			if sourceID == source.ID {
				return
			}
		}

		nodes.MustStoreGitSource(source)

		user.SourceIDs = append(user.SourceIDs, source.ID)
		nodes.MustStoreUser(user)

		c.GitSources = append(
			c.GitSources,
			GitSourceConfig{
				Repository: input.Repository,
				Reference:  input.Reference,
				ID:         source.ID,
			},
		)

		subs.Publish(SourceUpserted, source.ID)
	})

	return source.ID
}

// DeleteSource deletes a source.
func (c *SourcesConfig) DeleteSource(
	nodes *NodeManager,
	subs *pubsub.PubSub,
	userID string,
	id string,
) error {
	return nodes.LockUserE(userID, func(user User) error {
		parts, err := relay.DecodeID(id)
		if err != nil {
			return err
		}

		for i, v := range user.SourceIDs {
			if v == id {
				user.SourceIDs = append(
					user.SourceIDs[:i],
					user.SourceIDs[i+1:]...,
				)
				break
			}
		}

		switch parts[0] {
		case NodeTypeDirectorySource:
			// We can't delete the actual node because other node might reference it.
			for i, v := range c.DirectorySources {
				if v.ID == id {
					c.DirectorySources = append(
						c.DirectorySources[:i],
						c.DirectorySources[i+1:]...,
					)
					break
				}
			}
		case NodeTypeGitSource:
			// We can't delete the actual node because other node might reference it.
			for i, v := range c.GitSources {
				if v.ID == id {
					c.GitSources = append(
						c.GitSources[:i],
						c.GitSources[i+1:]...,
					)
					break
				}
			}
		default:
			return ErrType
		}

		nodes.MustStoreUser(user)
		subs.Publish(SourceDeleted, id)

		return nil
	})
}

// Save saves the config to disk, overwriting the file if it exists.
func (c SourcesConfig) Save() error {
	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(c.Filename), 0755); err != nil {
		return err
	}

	return ioutil.WriteFile(c.Filename, bytes, 0644)
}

// LoadSourcesConfigYAML loads a source config from a YAML file.
// It will create a file if it doesn't exist.
func LoadSourcesConfigYAML(filename string) (*SourcesConfig, error) {
	config := SourcesConfig{
		Filename: filename,
	}

	bytes, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		config := SourcesConfig{
			Filename: filename,
		}
		if err := config.Save(); err != nil {
			return nil, err
		}

		return LoadSourcesConfigYAML(filename)
	}
	if err != nil {
		return nil, err
	}

	err = yaml.UnmarshalStrict(bytes, &config)

	return &config, err
}

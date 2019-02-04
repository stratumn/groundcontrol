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

	yaml "gopkg.in/yaml.v2"

	"github.com/stratumn/groundcontrol/relay"
)

// SourceConfig contains all the data in a YAML source config file.
type SourceConfig struct {
	Filename         string `json:"filename"`
	DirectorySources []struct {
		Directory string `json:"directory"`
	} `json:"directorySources"`
	GitSources []struct {
		Repository string `json:"repository"`
		Branch     string `json:"branch"`
	} `json:"gitSources"`
}

// UpsertNodes upserts nodes for the content of the source config.
// The user node must already exists
func (c SourceConfig) UpsertNodes(nodes *NodeManager, userID string) error {
	var sourceIDs []string

	for _, sourceConfig := range c.DirectorySources {
		source := DirectorySource{
			ID:        relay.EncodeID(NodeTypeDirectorySource, sourceConfig.Directory),
			Directory: sourceConfig.Directory,
		}

		nodes.MustStoreDirectorySource(source)
		sourceIDs = append(sourceIDs, source.ID)
	}

	for _, sourceConfig := range c.GitSources {
		source := GitSource{
			ID: relay.EncodeID(
				NodeTypeDirectorySource,
				sourceConfig.Repository,
				sourceConfig.Branch,
			),
			Repository: sourceConfig.Repository,
			Branch:     sourceConfig.Branch,
		}

		nodes.MustStoreGitSource(source)
		sourceIDs = append(sourceIDs, source.ID)
	}

	nodes.MustLockUser(userID, func(user User) {
		user.SourceIDs = sourceIDs
		nodes.MustStoreUser(user)
	})

	return nil
}

// LoadSourceConfigYAML loads a source config from a YAML file.
func LoadSourceConfigYAML(filename string) (SourceConfig, error) {
	config := SourceConfig{
		Filename: filename,
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = yaml.UnmarshalStrict(bytes, &config)

	return config, err
}

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

// Config contains all the data in a YAML config file.
type Config struct {
	Filename   string `json:"filename"`
	Workspaces []struct {
		Name        string  `json:"name"`
		Slug        string  `json:"slug"`
		Description string  `json:"description"`
		Notes       *string `json:"notes"`
		Projects    []struct {
			Repository  string  `json:"repository"`
			Branch      string  `json:"branch"`
			Description *string `json:"description"`
		} `json:"projects"`
		Tasks []struct {
			Name  string `json:"name"`
			Steps []struct {
				Projects   []string `json:"projects"`
				Commands   []string `json:"commands"`
				Background bool     `json:"background"`
			} `json:"tasks"`
		} `json:"tasks"`
	} `json:"workspaces"`
}

// CreateNodes creates nodes for the content of the config.
// It returns a user node with has an ID derived from the filename of the config.
func (c Config) CreateNodes(nodes *NodeManager) (User, error) {
	user := User{
		ID: relay.EncodeID(NodeTypeUser, c.Filename),
	}

	for _, workspaceConfig := range c.Workspaces {
		workspace := Workspace{
			ID:          relay.EncodeID(NodeTypeWorkspace, workspaceConfig.Slug),
			Name:        workspaceConfig.Name,
			Slug:        workspaceConfig.Slug,
			Description: workspaceConfig.Description,
			Notes:       workspaceConfig.Notes,
		}

		for _, projectConfig := range workspaceConfig.Projects {
			project := Project{
				ID: relay.EncodeID(
					NodeTypeProject,
					workspace.Slug,
					projectConfig.Repository,
					projectConfig.Branch,
				),
				Repository:  projectConfig.Repository,
				Branch:      projectConfig.Branch,
				Description: nil,
				WorkspaceID: workspace.ID,
			}

			nodes.MustStoreProject(project)
			workspace.ProjectIDs = append(workspace.ProjectIDs, project.ID)
		}

		nodes.MustStoreWorkspace(workspace)
		user.WorkspaceIDs = append(user.WorkspaceIDs, workspace.ID)
	}

	return user, nodes.StoreUser(user)
}

// LoadConfigYAML loads a config from a YAML file.
func LoadConfigYAML(filename string) (Config, error) {
	config := Config{
		Filename: filename,
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = yaml.UnmarshalStrict(bytes, &config)

	return config, err
}

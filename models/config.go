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
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"github.com/stratumn/groundcontrol/relay"
)

// Config contains all the data in a YAML config file.
type Config struct {
	Filename   string `json:"filename"`
	Workspaces []struct {
		Slug        string  `json:"slug"`
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Notes       *string `json:"notes"`
		Projects    []struct {
			Slug        string  `json:"slug"`
			Repository  string  `json:"repository"`
			Branch      string  `json:"branch"`
			Description *string `json:"description"`
		} `json:"projects"`
		Tasks []struct {
			Name  string `json:"name"`
			Steps []struct {
				Projects []string `json:"projects"`
				Commands []string `json:"commands"`
			} `json:"tasks"`
		} `json:"tasks"`
	} `json:"workspaces"`
}

// CreateNodes creates nodes for the content of the config.
// The user node must already exists
func (c Config) CreateNodes(nodes *NodeManager, userID string) error {
	var workspaceIDs []string

	for _, workspaceConfig := range c.Workspaces {
		workspace := Workspace{
			ID:          relay.EncodeID(NodeTypeWorkspace, workspaceConfig.Slug),
			Slug:        workspaceConfig.Slug,
			Name:        workspaceConfig.Name,
			Description: workspaceConfig.Description,
			Notes:       workspaceConfig.Notes,
		}

		projectSlugToID := map[string]string{}

		for _, projectConfig := range workspaceConfig.Projects {
			project := Project{
				ID: relay.EncodeID(
					NodeTypeProject,
					workspace.Slug,
					projectConfig.Slug,
				),
				Slug:        projectConfig.Slug,
				Repository:  projectConfig.Repository,
				Branch:      projectConfig.Branch,
				Description: projectConfig.Description,
				WorkspaceID: workspace.ID,
			}

			nodes.MustStoreProject(project)
			workspace.ProjectIDs = append(workspace.ProjectIDs, project.ID)
			projectSlugToID[project.Slug] = project.ID
		}

		for i, taskConfig := range workspaceConfig.Tasks {
			task := Task{
				ID: relay.EncodeID(
					NodeTypeTask,
					workspace.Slug,
					fmt.Sprint(i),
				),
				Name:        taskConfig.Name,
				WorkspaceID: workspace.ID,
			}

			for j, stepConfig := range taskConfig.Steps {
				var projectIDs []string

				for _, slug := range stepConfig.Projects {
					id, ok := projectSlugToID[slug]
					if !ok {
						return ErrNotFound
					}
					projectIDs = append(projectIDs, id)
				}

				step := Step{
					ID: relay.EncodeID(
						NodeTypeStep,
						workspace.Slug,
						fmt.Sprint(i),
						fmt.Sprint(j),
					),
					ProjectIDs: projectIDs,
					Commands:   stepConfig.Commands,
					TaskID:     task.ID,
				}

				nodes.MustStoreStep(step)
				task.StepIDs = append(task.StepIDs, step.ID)
			}

			nodes.MustStoreTask(task)
			workspace.TaskIDs = append(workspace.TaskIDs, task.ID)
		}

		nodes.MustStoreWorkspace(workspace)
		workspaceIDs = append(workspaceIDs, workspace.ID)
	}

	nodes.MustLockUser(userID, func(user User) {
		user.WorkspaceIDs = append(user.WorkspaceIDs, workspaceIDs...)
		nodes.MustStoreUser(user)
	})

	return nil
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

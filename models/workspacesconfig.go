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

	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/relay"
)

// WorkspacesConfig contains all the data in a YAML workspaces config file.
type WorkspacesConfig struct {
	Filename   string            `json:"-" yaml:"-"`
	Workspaces []WorkspaceConfig `json:"workspaces"`
}

// WorkspaceConfig contains all the data in a YAML workspace config file.
type WorkspaceConfig struct {
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Notes       *string         `json:"notes"`
	Projects    []ProjectConfig `json:"projects" yaml:",flow"`
	Tasks       []TaskConfig    `json:"tasks"`
}

// ProjectConfig contains all the data in a YAML project config file.
type ProjectConfig struct {
	Slug        string  `json:"slug"`
	Repository  string  `json:"repository"`
	Branch      string  `json:"branch"`
	Description *string `json:"description"`
}

// TaskConfig contains all the data in a YAML task config file.
type TaskConfig struct {
	Name  string       `json:"name"`
	Steps []StepConfig `json:"tasks"`
}

// StepConfig contains all the data in a YAML step config file.
type StepConfig struct {
	Projects []string `json:"projects"`
	Commands []string `json:"commands"`
}

// UpsertNodes upserts nodes for the content of the config.
// It returns the IDs of the workspaces upserted.
func (c WorkspacesConfig) UpsertNodes(
	nodes *NodeManager,
	subs *pubsub.PubSub,
) ([]string, error) {
	var workspaceIDs []string

	for _, workspaceConfig := range c.Workspaces {
		id, err := workspaceConfig.UpsertNodes(nodes, subs)
		if err != nil {
			return nil, err
		}

		workspaceIDs = append(workspaceIDs, id)
	}

	return workspaceIDs, nil
}

// UpsertNodes upserts nodes for the content of the config.
// It returns the ID of the workspace upserted.
func (c WorkspaceConfig) UpsertNodes(
	nodes *NodeManager,
	subs *pubsub.PubSub,
) (string, error) {
	id := relay.EncodeID(NodeTypeWorkspace, c.Slug)

	err := nodes.MustLockOrNewWorkspaceE(id, func(workspace Workspace) error {
		workspace.Slug = c.Slug
		workspace.Name = c.Name
		workspace.Description = c.Description
		workspace.Notes = c.Notes
		workspace.ProjectIDs = nil
		workspace.TaskIDs = nil
		projectSlugToID := map[string]string{}

		for _, projectConfig := range c.Projects {
			projectID := projectConfig.UpsertNodes(nodes, subs, id, c.Slug)
			workspace.ProjectIDs = append(workspace.ProjectIDs, projectID)
			projectSlugToID[projectConfig.Slug] = projectID
		}

		for _, taskConfig := range c.Tasks {
			taskID, err := taskConfig.UpsertNodes(nodes, subs, id, workspace.Slug, projectSlugToID)
			if err != nil {
				return err
			}

			workspace.TaskIDs = append(workspace.TaskIDs, taskID)
		}

		nodes.MustStoreWorkspace(workspace)
		subs.Publish(WorkspaceUpserted, id)

		return nil
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

// UpsertNodes upserts nodes for the content of the config.
// It returns the ID of the project upserted.
func (c ProjectConfig) UpsertNodes(
	nodes *NodeManager,
	subs *pubsub.PubSub,
	workspaceID string,
	workspaceSlug string,
) string {
	id := relay.EncodeID(
		NodeTypeProject,
		workspaceSlug,
		c.Slug,
	)

	nodes.MustLockOrNewProject(id, func(project Project) {
		project.Slug = c.Slug
		project.Repository = c.Repository
		project.Branch = c.Branch
		project.Description = c.Description
		project.WorkspaceID = workspaceID

		nodes.MustStoreProject(project)
		subs.Publish(ProjectUpserted, id)
	})

	return id
}

// UpsertNodes upserts nodes for the content of the config.
// It returns the ID of the task upserted.
func (c TaskConfig) UpsertNodes(
	nodes *NodeManager,
	subs *pubsub.PubSub,
	workspaceID string,
	workspaceSlug string,
	projectSlugToID map[string]string,
) (string, error) {
	id := relay.EncodeID(
		NodeTypeTask,
		workspaceSlug,
		c.Name,
	)

	err := nodes.MustLockOrNewTaskE(id, func(task Task) error {
		task.Name = c.Name
		task.WorkspaceID = workspaceID
		task.StepIDs = nil

		for stepIndex, stepConfig := range c.Steps {
			stepID, err := stepConfig.UpsertNodes(
				nodes,
				workspaceSlug,
				id,
				c.Name,
				stepIndex,
				projectSlugToID,
			)
			if err != nil {
				return err
			}

			task.StepIDs = append(task.StepIDs, stepID)
		}

		nodes.MustStoreTask(task)
		subs.Publish(TaskUpserted, id)

		return nil
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

// UpsertNodes upserts nodes for the content of the config.
// It returns the ID of the step upserted.
func (c StepConfig) UpsertNodes(
	nodes *NodeManager,
	workspaceSlug string,
	taskID string,
	taskName string,
	stepIndex int,
	projectSlugToID map[string]string,
) (string, error) {
	id := relay.EncodeID(
		NodeTypeStep,
		workspaceSlug,
		taskName,
		fmt.Sprint(stepIndex),
	)

	err := nodes.MustLockOrNewStepE(id, func(step Step) error {
		step.TaskID = taskID
		step.ProjectIDs = nil
		step.CommandIDs = nil

		for _, slug := range c.Projects {
			id, ok := projectSlugToID[slug]
			if !ok {
				return ErrNotFound
			}
			step.ProjectIDs = append(step.ProjectIDs, id)
		}

		for commandIndex, command := range c.Commands {
			id := relay.EncodeID(
				NodeTypeCommand,
				workspaceSlug,
				taskName,
				fmt.Sprint(stepIndex),
				fmt.Sprint(commandIndex),
			)
			nodes.MustStoreCommand(Command{
				ID:      id,
				Command: command,
			})
			step.CommandIDs = append(step.CommandIDs, id)
		}

		nodes.MustStoreStep(step)

		return nil
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

// LoadWorkspacesConfigYAML loads a config from a YAML file.
func LoadWorkspacesConfigYAML(filename string) (WorkspacesConfig, error) {
	config := WorkspacesConfig{
		Filename: filename,
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = yaml.UnmarshalStrict(bytes, &config)

	return config, err
}

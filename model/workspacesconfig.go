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
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"

	"groundcontrol/relay"
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
	Description *string         `json:"description"`
	Notes       *string         `json:"notes"`
	Projects    []ProjectConfig `json:"projects" yaml:",flow"`
	Tasks       []TaskConfig    `json:"tasks"`
}

// ProjectConfig contains all the data in a YAML project config file.
type ProjectConfig struct {
	Slug        string  `json:"slug"`
	Repository  string  `json:"repository"`
	Reference   string  `json:"reference"`
	Description *string `json:"description"`
}

// TaskConfig contains all the data in a YAML task config file.
type TaskConfig struct {
	Name      string           `json:"name"`
	Variables []VariableConfig `json:"variables"`
	Steps     []StepConfig     `json:"tasks"`
}

// VariableConfig contains all the data in a YAML variable config file.
type VariableConfig struct {
	Name    string  `json:"name"`
	Default *string `json:"default"`
}

// StepConfig contains all the data in a YAML step config file.
type StepConfig struct {
	Projects []string `json:"projects"`
	Commands []string `json:"commands"`
}

// UpsertNodes upserts nodes for the content of the config.
// It returns the IDs of the workspaces upserted.
func (c WorkspacesConfig) UpsertNodes(ctx context.Context, sourceID string) ([]string, error) {
	var workspaceIDs []string

	for _, workspaceConfig := range c.Workspaces {
		id, err := workspaceConfig.UpsertNodes(ctx, sourceID)
		if err != nil {
			return nil, err
		}

		workspaceIDs = append(workspaceIDs, id)
	}

	return workspaceIDs, nil
}

// UpsertNodes upserts nodes for the content of the config.
// It returns the ID of the workspace upserted.
func (c WorkspaceConfig) UpsertNodes(ctx context.Context, sourceID string) (string, error) {
	id := relay.EncodeID(NodeTypeWorkspace, c.Slug)

	err := MustLockOrNewWorkspaceE(ctx, id, func(workspace *Workspace) error {
		workspace.Slug = c.Slug
		workspace.Name = c.Name
		workspace.Description = c.Description
		workspace.Notes = c.Notes
		workspace.SourceID = sourceID
		workspace.ProjectsIDs = nil
		workspace.TasksIDs = nil
		projectSlugToID := map[string]string{}

		// We need to make sure the workspace exists before child nodes refer to it.
		workspace.MustStore(ctx)

		for _, projectConfig := range c.Projects {
			projectID := projectConfig.UpsertNodes(ctx, id, c.Slug)
			workspace.ProjectsIDs = append(workspace.ProjectsIDs, projectID)
			projectSlugToID[projectConfig.Slug] = projectID
		}

		for _, taskConfig := range c.Tasks {
			taskID, err := taskConfig.UpsertNodes(ctx, id, workspace.Slug, projectSlugToID)
			if err != nil {
				return err
			}

			workspace.TasksIDs = append(workspace.TasksIDs, taskID)
		}

		workspace.MustStore(ctx)

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
	ctx context.Context,
	workspaceID string,
	workspaceSlug string,
) string {
	id := relay.EncodeID(
		NodeTypeProject,
		workspaceSlug,
		c.Slug,
	)

	MustLockOrNewProject(ctx, id, func(project *Project) {
		project.Slug = c.Slug
		project.Repository = c.Repository
		project.Reference = c.Reference
		project.Description = c.Description
		project.WorkspaceID = workspaceID

		if _, err := LoadProject(ctx, id); err == ErrNotFound {
			project.RemoteReference = c.Reference
			project.LocalReference = c.Reference
			project.IsClean = true
		}

		project.MustStore(ctx)
	})

	return id
}

// UpsertNodes upserts nodes for the content of the config.
// It returns the ID of the task upserted.
func (c TaskConfig) UpsertNodes(
	ctx context.Context,
	workspaceID string,
	workspaceSlug string,
	projectSlugToID map[string]string,
) (string, error) {
	id := relay.EncodeID(
		NodeTypeTask,
		workspaceSlug,
		c.Name,
	)

	err := MustLockOrNewTaskE(ctx, id, func(task *Task) error {
		task.Name = c.Name
		task.WorkspaceID = workspaceID
		task.VariablesIDs = nil
		task.StepsIDs = nil

		// We need to make sure the task exists before child nodes refer to it.
		task.MustStore(ctx)

		for variableIndex, variableConfig := range c.Variables {
			variableID := variableConfig.UpsertNodes(
				ctx,
				workspaceSlug,
				id,
				c.Name,
				variableIndex,
			)

			task.VariablesIDs = append(task.VariablesIDs, variableID)
		}

		for stepIndex, stepConfig := range c.Steps {
			stepID, err := stepConfig.UpsertNodes(
				ctx,
				workspaceSlug,
				id,
				c.Name,
				stepIndex,
				projectSlugToID,
			)
			if err != nil {
				return err
			}

			task.StepsIDs = append(task.StepsIDs, stepID)
		}

		task.MustStore(ctx)

		return nil
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

// UpsertNodes upserts nodes for the content of the config.
// It returns the ID of the variable upserted.
func (c VariableConfig) UpsertNodes(
	ctx context.Context,
	workspaceSlug string,
	taskID string,
	taskName string,
	stepIndex int,
) string {
	id := relay.EncodeID(
		NodeTypeVariable,
		workspaceSlug,
		taskName,
		fmt.Sprint(stepIndex),
	)

	MustLockOrNewVariable(ctx, id, func(variable *Variable) {
		variable.Name = c.Name
		variable.Default = c.Default

		variable.MustStore(ctx)
	})

	return id
}

// UpsertNodes upserts nodes for the content of the config.
// It returns the ID of the step upserted.
func (c StepConfig) UpsertNodes(
	ctx context.Context,
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

	err := MustLockOrNewStepE(ctx, id, func(step *Step) error {
		step.TaskID = taskID
		step.ProjectsIDs = nil
		step.CommandsIDs = nil

		for _, slug := range c.Projects {
			id, ok := projectSlugToID[slug]
			if !ok {
				return ErrNotFound
			}
			step.ProjectsIDs = append(step.ProjectsIDs, id)
		}

		for commandIndex, command := range c.Commands {
			id := relay.EncodeID(
				NodeTypeCommand,
				workspaceSlug,
				taskName,
				fmt.Sprint(stepIndex),
				fmt.Sprint(commandIndex),
			)
			(&Command{ID: id, Command: command}).MustStore(ctx)
			step.CommandsIDs = append(step.CommandsIDs, id)
		}

		step.MustStore(ctx)

		return nil
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

// LoadWorkspacesConfigYAML loads a config from a YAML file.
func LoadWorkspacesConfigYAML(filename string) (*WorkspacesConfig, error) {
	config := WorkspacesConfig{
		Filename: filename,
	}

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return &config, yaml.UnmarshalStrict(bytes, &config)
}

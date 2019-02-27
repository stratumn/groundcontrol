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
	Services    []ServiceConfig `json:"services"`
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

// ServiceConfig contains all the data in a YAML service config file.
type ServiceConfig struct {
	Name      string           `json:"name"`
	Variables []VariableConfig `json:"variables"`
	Project   string           `json:"project"`
	Needs     []string         `json:"needs"`
	Command   string           `json:"command"`
	Before    []string         `json:"before"`
	After     []string         `json:"after"`
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

	err := MustLockOrNewWorkspaceE(ctx, id, func(workspace *Workspace, isNew bool) error {
		workspace.Slug = c.Slug
		workspace.Name = c.Name
		workspace.Description = c.Description
		workspace.Notes = c.Notes
		workspace.SourceID = sourceID
		workspace.ProjectsIDs = nil
		workspace.TasksIDs = nil
		workspace.ServicesIDs = nil
		projectSlugToID := map[string]string{}
		taskNameToID := map[string]string{}
		serviceNameToID := map[string]string{}

		if isNew {
			// We need to make sure the workspace exists before child nodes refer to it.
			workspace.MustStore(ctx)
		}

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
			taskNameToID[taskConfig.Name] = taskID
		}

		for _, serviceConfig := range c.Services {
			serviceID, err := serviceConfig.UpsertNodes(
				ctx,
				id,
				workspace.Slug,
				projectSlugToID,
				taskNameToID,
			)
			if err != nil {
				return err
			}

			workspace.ServicesIDs = append(workspace.ServicesIDs, serviceID)
			serviceNameToID[serviceConfig.Name] = serviceID
		}

		for _, serviceConfig := range c.Services {
			err := serviceConfig.SetNeeds(ctx, workspace.Slug, serviceNameToID)
			if err != nil {
				return err
			}
		}

		for _, serviceConfig := range c.Services {
			err := serviceConfig.SetDependencies(ctx, workspace.Slug)
			if err != nil {
				return err
			}
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

	MustLockOrNewProject(ctx, id, func(project *Project, isNew bool) {
		project.Slug = c.Slug
		project.Repository = c.Repository
		project.Reference = c.Reference
		project.Description = c.Description
		project.WorkspaceID = workspaceID

		if isNew {
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

	err := MustLockOrNewTaskE(ctx, id, func(task *Task, isNew bool) error {
		task.Name = c.Name
		task.WorkspaceID = workspaceID
		task.VariablesIDs = nil
		task.StepsIDs = nil

		if isNew {
			// We need to make sure the task exists before child nodes refer to it.
			task.MustStore(ctx)
		}

		for _, variableConfig := range c.Variables {
			variableID := variableConfig.UpsertNodes(
				ctx,
				workspaceSlug,
				c.Name,
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
	namespace string,
) string {
	id := relay.EncodeID(
		NodeTypeVariable,
		workspaceSlug,
		namespace,
		c.Name,
	)

	MustLockOrNewVariable(ctx, id, func(variable *Variable, _ bool) {
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

	err := MustLockOrNewStepE(ctx, id, func(step *Step, _ bool) error {
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

// UpsertNodes upserts nodes for the content of the config.
// It returns the ID of the service upserted.
func (c ServiceConfig) UpsertNodes(
	ctx context.Context,
	workspaceID string,
	workspaceSlug string,
	projectSlugToID map[string]string,
	taskNameToID map[string]string,
) (string, error) {
	id := relay.EncodeID(
		NodeTypeService,
		workspaceSlug,
		c.Name,
	)

	err := MustLockOrNewServiceE(ctx, id, func(service *Service, isNew bool) error {
		service.Name = c.Name
		service.WorkspaceID = workspaceID
		service.Command = c.Command
		service.VariablesIDs = nil
		service.NeedsIDs = nil
		service.BeforeIDs = nil
		service.AfterIDs = nil

		if isNew {
			service.Status = ServiceStatusStopped
		}

		if c.Project != "" {
			projectID, ok := projectSlugToID[c.Project]
			if !ok {
				return ErrNotFound
			}

			service.ProjectID = projectID
		}

		for _, variableConfig := range c.Variables {
			variableID := variableConfig.UpsertNodes(
				ctx,
				workspaceSlug,
				c.Name,
			)

			service.VariablesIDs = append(service.VariablesIDs, variableID)
		}

		for _, taskName := range c.Before {
			taskID, ok := taskNameToID[taskName]
			if !ok {
				return ErrNotFound
			}

			service.BeforeIDs = append(service.BeforeIDs, taskID)
		}

		for _, taskName := range c.After {
			taskID, ok := taskNameToID[taskName]
			if !ok {
				return ErrNotFound
			}

			service.AfterIDs = append(service.AfterIDs, taskID)
		}

		service.MustStore(ctx)

		return nil
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

// SetNeeds sets nodes with the Services it needs.
// It must be called after all the Service nodes have been created.
func (c ServiceConfig) SetNeeds(
	ctx context.Context,
	workspaceSlug string,
	serviceNameToID map[string]string,
) error {
	id := relay.EncodeID(
		NodeTypeService,
		workspaceSlug,
		c.Name,
	)

	return MustLockServiceE(ctx, id, func(service *Service) error {
		service.NeedsIDs = nil

		for _, serviceName := range c.Needs {
			serviceID, ok := serviceNameToID[serviceName]
			if !ok {
				return ErrNotFound
			}

			service.NeedsIDs = append(service.NeedsIDs, serviceID)
		}

		service.MustStore(ctx)

		return nil
	})
}

// SetDependencies sets nodes with the Services and Variables it depends on.
// It must be called after SetNeeds has been called on all the Services of the Workspace.
func (c ServiceConfig) SetDependencies(ctx context.Context, workspaceSlug string) error {
	id := relay.EncodeID(
		NodeTypeService,
		workspaceSlug,
		c.Name,
	)

	return MustLockServiceE(ctx, id, func(service *Service) error {
		if err := service.ComputeDependencies(ctx); err != nil {
			return err
		}

		service.ComputeAllVariables(ctx)
		service.MustStore(ctx)

		return nil
	})
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

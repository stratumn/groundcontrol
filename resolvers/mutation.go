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

package resolvers

import (
	"context"

	"github.com/stratumn/groundcontrol/jobs"
	"github.com/stratumn/groundcontrol/models"
)

type mutationResolver struct {
	*Resolver
}

func (r *mutationResolver) AddDirectorySource(
	ctx context.Context,
	input models.DirectorySourceInput,
) (models.DirectorySource, error) {
	modelCtx := models.GetModelContext(ctx)

	id := modelCtx.Sources.UpsertDirectorySource(
		modelCtx.Nodes,
		modelCtx.Subs,
		modelCtx.ViewerID,
		input,
	)

	return modelCtx.Nodes.MustLoadDirectorySource(id), nil
}

func (r *mutationResolver) AddGitSource(
	ctx context.Context,
	input models.GitSourceInput,
) (models.GitSource, error) {
	modelCtx := models.GetModelContext(ctx)

	id := modelCtx.Sources.UpsertGitSource(
		modelCtx.Nodes,
		modelCtx.Subs,
		modelCtx.ViewerID,
		input,
	)

	return modelCtx.Nodes.MustLoadGitSource(id), nil
}

func (r *mutationResolver) LoadProjectCommits(ctx context.Context, id string) (models.Job, error) {
	nodes := models.GetModelContext(ctx).Nodes

	jobID, err := jobs.LoadCommits(ctx, id, models.JobPriorityHigh)
	if err != nil {
		return models.Job{}, err
	}

	return nodes.MustLoadJob(jobID), nil
}

func (r *mutationResolver) LoadWorkspaceCommits(ctx context.Context, id string) ([]models.Job, error) {
	nodes := models.GetModelContext(ctx).Nodes

	workspace, err := nodes.LoadWorkspace(id)
	if err != nil {
		return nil, err
	}

	var slice []models.Job

	for _, projectID := range workspace.ProjectIDs {
		project := nodes.MustLoadProject(projectID)

		if project.IsLoadingCommits || len(project.CommitIDs) > 0 {
			continue
		}

		jobID, err := jobs.LoadCommits(ctx, project.ID, models.JobPriorityHigh)
		if err != nil {
			return nil, err
		}

		slice = append(slice, nodes.MustLoadJob(jobID))
	}

	return slice, nil
}

func (r *mutationResolver) CloneProject(ctx context.Context, id string) (models.Job, error) {
	nodes := models.GetModelContext(ctx).Nodes

	jobID, err := jobs.Clone(ctx, id, models.JobPriorityHigh)
	if err != nil {
		return models.Job{}, err
	}

	return nodes.MustLoadJob(jobID), nil
}

func (r *mutationResolver) CloneWorkspace(ctx context.Context, id string) ([]models.Job, error) {
	nodes := models.GetModelContext(ctx).Nodes

	workspace, err := nodes.LoadWorkspace(id)
	if err != nil {
		return nil, err
	}

	var slice []models.Job

	for _, projectID := range workspace.ProjectIDs {
		project := nodes.MustLoadProject(projectID)

		if project.IsCloning || project.IsCloned(ctx) {
			continue
		}

		jobID, err := jobs.Clone(ctx, project.ID, models.JobPriorityHigh)
		if err != nil {
			return nil, err
		}

		slice = append(slice, nodes.MustLoadJob(jobID))
	}

	return slice, nil
}

func (r *mutationResolver) PullProject(ctx context.Context, id string) (models.Job, error) {
	nodes := models.GetModelContext(ctx).Nodes

	jobID, err := jobs.Pull(ctx, id, models.JobPriorityHigh)
	if err != nil {
		return models.Job{}, err
	}

	return nodes.MustLoadJob(jobID), nil
}

func (r *mutationResolver) PullWorkspace(ctx context.Context, id string) ([]models.Job, error) {
	nodes := models.GetModelContext(ctx).Nodes

	workspace, err := nodes.LoadWorkspace(id)
	if err != nil {
		return nil, err
	}

	var slice []models.Job

	for _, projectID := range workspace.ProjectIDs {
		project := nodes.MustLoadProject(projectID)

		if project.IsPulling || !project.IsCloned(ctx) || !project.IsBehind {
			continue
		}

		jobID, err := jobs.Pull(ctx, project.ID, models.JobPriorityHigh)
		if err != nil {
			return nil, err
		}

		slice = append(slice, nodes.MustLoadJob(jobID))
	}

	return slice, nil
}

func (r *mutationResolver) Run(ctx context.Context, id string) (models.Job, error) {
	nodes := models.GetModelContext(ctx).Nodes

	jobID, err := jobs.Run(ctx, id, models.JobPriorityHigh)
	if err != nil {
		return models.Job{}, err
	}

	return nodes.MustLoadJob(jobID), nil
}

func (r *mutationResolver) StopJob(ctx context.Context, id string) (models.Job, error) {
	modelCtx := models.GetModelContext(ctx)
	jobs := modelCtx.Jobs
	nodes := modelCtx.Nodes

	if err := jobs.Stop(modelCtx, id); err != nil {
		return models.Job{}, nil
	}

	return nodes.MustLoadJob(id), nil
}

func (r *mutationResolver) StartProcessGroup(ctx context.Context, id string) (models.ProcessGroup, error) {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	pm := modelCtx.PM

	processGroup, err := nodes.LoadProcessGroup(id)
	if err != nil {
		return models.ProcessGroup{}, err
	}

	for _, processID := range processGroup.ProcessIDs {
		process := nodes.MustLoadProcess(processID)

		if process.Status == models.ProcessStatusRunning {
			continue
		}

		if err := pm.Start(ctx, process.ID); err != nil {
			return models.ProcessGroup{}, err
		}
	}

	return processGroup, nil
}

func (r *mutationResolver) StopProcessGroup(ctx context.Context, id string) (models.ProcessGroup, error) {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	pm := modelCtx.PM

	processGroup, err := nodes.LoadProcessGroup(id)
	if err != nil {
		return models.ProcessGroup{}, err
	}

	for _, processID := range processGroup.ProcessIDs {
		process := nodes.MustLoadProcess(processID)

		if process.Status != models.ProcessStatusRunning {
			continue
		}

		if err := pm.Stop(ctx, process.ID); err != nil {
			return models.ProcessGroup{}, err
		}
	}

	return processGroup, nil
}

func (r *mutationResolver) StartProcess(ctx context.Context, id string) (models.Process, error) {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	pm := modelCtx.PM

	err := pm.Start(ctx, id)
	if err != nil {
		return models.Process{}, err
	}

	return nodes.MustLoadProcess(id), nil
}

func (r *mutationResolver) StopProcess(ctx context.Context, id string) (models.Process, error) {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	pm := modelCtx.PM

	err := pm.Stop(ctx, id)
	if err != nil {
		return models.Process{}, err
	}

	return nodes.MustLoadProcess(id), nil
}

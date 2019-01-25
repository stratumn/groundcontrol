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

func (r *mutationResolver) LoadProjectCommits(ctx context.Context, id string) (models.Job, error) {
	jobID, err := jobs.LoadCommits(r.Nodes, r.Jobs, r.Subs, id)
	if err != nil {
		return models.Job{}, err
	}

	return r.Nodes.MustLoadJob(jobID), nil
}

func (r *mutationResolver) LoadWorkspaceCommits(ctx context.Context, id string) ([]models.Job, error) {
	workspace, err := r.Nodes.LoadWorkspace(id)
	if err != nil {
		return nil, err
	}

	var slice []models.Job

	for _, project := range workspace.Projects(r.Nodes) {
		if project.IsLoadingCommits || len(project.CommitIDs) > 0 {
			continue
		}

		jobID, err := jobs.LoadCommits(r.Nodes, r.Jobs, r.Subs, project.ID)
		if err != nil {
			return nil, err
		}

		slice = append(slice, r.Nodes.MustLoadJob(jobID))
	}

	return slice, nil
}

func (r *mutationResolver) CloneProject(ctx context.Context, id string) (models.Job, error) {
	jobID, err := jobs.Clone(r.Nodes, r.Jobs, r.Subs, r.GetProjectPath, id)
	if err != nil {
		return models.Job{}, err
	}

	return r.Nodes.MustLoadJob(jobID), nil
}

func (r *mutationResolver) CloneWorkspace(ctx context.Context, id string) ([]models.Job, error) {
	workspace, err := r.Nodes.LoadWorkspace(id)
	if err != nil {
		return nil, err
	}

	var slice []models.Job

	for _, project := range workspace.Projects(r.Nodes) {
		if project.IsCloning || project.IsCloned(r.Nodes, r.GetProjectPath) {
			continue
		}

		jobID, err := jobs.Clone(r.Nodes, r.Jobs, r.Subs, r.GetProjectPath, project.ID)
		if err != nil {
			return nil, err
		}

		slice = append(slice, r.Nodes.MustLoadJob(jobID))
	}

	return slice, nil
}

func (r *mutationResolver) PullProject(ctx context.Context, id string) (models.Job, error) {
	jobID, err := jobs.Pull(r.Nodes, r.Jobs, r.Subs, r.GetProjectPath, id)
	if err != nil {
		return models.Job{}, err
	}

	return r.Nodes.MustLoadJob(jobID), nil
}

func (r *mutationResolver) PullWorkspace(ctx context.Context, id string) ([]models.Job, error) {
	workspace, err := r.Nodes.LoadWorkspace(id)
	if err != nil {
		return nil, err
	}

	var slice []models.Job

	for _, project := range workspace.Projects(r.Nodes) {
		if project.IsPulling || !project.IsCloned(r.Nodes, r.GetProjectPath) {
			continue
		}

		jobID, err := jobs.Pull(r.Nodes, r.Jobs, r.Subs, r.GetProjectPath, project.ID)
		if err != nil {
			return nil, err
		}

		slice = append(slice, r.Nodes.MustLoadJob(jobID))
	}

	return slice, nil
}

func (r *mutationResolver) Run(ctx context.Context, id string) (models.Job, error) {
	jobID, err := jobs.Run(r.Nodes, r.Log, r.Jobs, r.Subs, r.GetProjectPath, id)
	if err != nil {
		return models.Job{}, err
	}

	return r.Nodes.MustLoadJob(jobID), nil
}

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

	"github.com/stratumn/groundcontrol/models"
)

type mutationResolver struct {
	*Resolver
}

func (r *mutationResolver) CloneProject(ctx context.Context, id string) (models.Job, error) {
	project, err := r.NodeManager.LoadProject(id)
	if err != nil {
		return models.Job{}, err
	}

	job, err := project.CloneJob(r.JobManager, r.PubSub, r.GetProjectPath)
	if err != nil {
		return models.Job{}, err
	}

	return *job, nil
}

func (r *mutationResolver) CloneWorkspace(ctx context.Context, id string) ([]models.Job, error) {
	workspace, err := r.NodeManager.LoadWorkspace(id)
	if err != nil {
		return nil, err
	}

	var jobs []models.Job

	for _, project := range workspace.Projects {
		job, err := project.CloneJob(r.JobManager, r.PubSub, r.GetProjectPath)
		if err == nil {
			jobs = append(jobs, *job)
			continue
		} else if err == models.ErrCloning || err == models.ErrCloned {
			continue
		}
		return jobs, err
	}

	return jobs, nil
}

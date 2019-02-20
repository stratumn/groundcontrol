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

package jobs

import (
	"context"

	"groundcontrol/models"
)

// LoadCommits loads the commits of a project from a remote repo and updates the project.
func LoadCommits(ctx context.Context, projectID string, priority models.JobPriority) (string, error) {
	if err := startLoadingCommits(ctx, projectID); err != nil {
		return "", err
	}

	modelCtx := models.GetModelContext(ctx)

	jobID := modelCtx.Jobs.Add(
		ctx,
		LoadCommitsJob,
		projectID,
		priority,
		func(ctx context.Context) error {
			return doLoadCommits(ctx, projectID)
		},
	)

	return jobID, nil
}

func startLoadingCommits(ctx context.Context, projectID string) error {
	modelCtx := models.GetModelContext(ctx)
	subs := modelCtx.Subs
	workspaceID := ""

	err := models.LockProjectE(ctx, projectID, func(project models.Project) error {
		if project.IsLoadingCommits {
			return ErrDuplicate
		}

		workspaceID = project.WorkspaceID
		project.IsLoadingCommits = true
		project.MustStore(ctx)

		return nil
	})

	if err != nil {
		return err
	}

	subs.Publish(models.ProjectUpserted, projectID)
	subs.Publish(models.WorkspaceUpserted, workspaceID)

	return nil
}

func doLoadCommits(ctx context.Context, projectID string) error {
	return models.MustLockProjectE(ctx, projectID, func(project models.Project) error {
		return project.Update(ctx)
	})
}

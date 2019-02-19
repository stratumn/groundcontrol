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

// Pull pulls a repository from origin.
func Pull(ctx context.Context, projectID string, priority models.JobPriority) (string, error) {
	if err := startPulling(ctx, projectID); err != nil {
		return "", err
	}

	modelCtx := models.GetModelContext(ctx)

	jobID := modelCtx.Jobs.Add(
		models.GetModelContext(ctx),
		LoadCommitsJob,
		projectID,
		priority,
		func(ctx context.Context) error {
			return doPull(ctx, projectID)
		},
	)

	return jobID, nil
}

func startPulling(ctx context.Context, projectID string) error {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	subs := modelCtx.Subs
	workspaceID := ""

	err := nodes.LockProjectE(projectID, func(project models.Project) error {
		if project.IsPulling {
			return ErrDuplicate
		}

		workspaceID = project.WorkspaceID
		project.IsPulling = true
		nodes.MustStoreProject(project)

		return nil
	})

	if err != nil {
		return err
	}

	subs.Publish(models.ProjectUpserted, projectID)
	subs.Publish(models.WorkspaceUpserted, workspaceID)

	return nil
}

func doPull(ctx context.Context, projectID string) error {
	nodes := models.GetModelContext(ctx).Nodes

	return nodes.MustLockProjectE(projectID, func(project models.Project) error {
		return project.Pull(ctx)
	})
}

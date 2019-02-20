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

// LoadAllCommits creates jobs to load the commits of every project.
// It doesn't return errors but will output a log message when errors happen.
func LoadAllCommits(ctx context.Context) []string {
	modelCtx := models.GetModelContext(ctx)
	viewer := models.MustLoadUser(ctx, modelCtx.ViewerID)

	var jobIDs []string

	for _, workspaceID := range viewer.WorkspaceIDs(ctx) {
		workspace := models.MustLoadWorkspace(ctx, workspaceID)

		for _, projectID := range workspace.ProjectIDs {
			project := models.MustLoadProject(ctx, projectID)

			if project.IsLoadingCommits {
				continue
			}

			jobID, err := LoadCommits(ctx, project.ID, models.JobPriorityNormal)
			if err != nil {
				modelCtx.Log.ErrorWithOwner(
					ctx,
					project.ID,
					"LoadCommits failed because %s",
					err.Error(),
				)
				continue
			}

			jobIDs = append(jobIDs, jobID)
		}
	}

	return jobIDs
}

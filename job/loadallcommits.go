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

package job

import (
	"context"

	"groundcontrol/model"
)

// LoadAllCommits creates jobs to load the commits of every project.
// It doesn't return errors but will output a log message when errors happen.
func LoadAllCommits(ctx context.Context, priority model.JobPriority) []string {
	modelCtx := model.GetModelContext(ctx)
	viewer := model.MustLoadUser(ctx, modelCtx.ViewerID)

	var jobIDs []string

	for _, workspaceID := range viewer.WorkspacesIDs(ctx) {
		workspace := model.MustLoadWorkspace(ctx, workspaceID)

		for _, projectID := range workspace.ProjectsIDs {
			project := model.MustLoadProject(ctx, projectID)

			if project.IsLoadingCommits {
				continue
			}

			jobID, err := LoadProjectCommits(ctx, project.ID, priority)
			if err != nil {
				modelCtx.Log.ErrorWithOwner(
					ctx,
					project.ID,
					"LoadProjectCommits failed because %s",
					err.Error(),
				)
				continue
			}

			jobIDs = append(jobIDs, jobID)
		}
	}

	return jobIDs
}

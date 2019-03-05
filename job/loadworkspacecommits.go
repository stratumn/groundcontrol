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

	"groundcontrol/appcontext"
	"groundcontrol/model"
)

// LoadWorkspaceCommits creates jobs to load the commits of every project in a workspace.
func LoadWorkspaceCommits(ctx context.Context, workspaceID string, highPriority bool) ([]string, error) {
	appCtx := appcontext.Get(ctx)
	workspace, err := model.LoadWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	var jobIDs []string

	for _, projectID := range workspace.ProjectsIDs {
		project := model.MustLoadProject(ctx, projectID)

		if project.IsLoadingCommits || len(project.RemoteCommitsIDs) > 0 {
			continue
		}

		jobID, err := LoadProjectCommits(ctx, project.ID, highPriority)
		if err != nil {
			appCtx.Log.ErrorWithOwner(
				ctx,
				appCtx.SystemID,
				"LoadProjectCommits failed because %s",
				err.Error(),
			)
			continue
		}

		jobIDs = append(jobIDs, jobID)
	}

	return jobIDs, nil
}

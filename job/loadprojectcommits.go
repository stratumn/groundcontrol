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

// LoadProjectCommits loads the commits of a project from a remote repo and updates the project.
func LoadProjectCommits(ctx context.Context, projectID string, highPriority bool) (string, error) {
	if err := startLoadingProjectCommits(ctx, projectID); err != nil {
		return "", err
	}

	modelCtx := model.GetContext(ctx)

	return modelCtx.Jobs.Add(
		ctx,
		JobNameLoadProjectCommits,
		projectID,
		highPriority,
		func(ctx context.Context) error {
			return doLoadProjectCommits(ctx, projectID)
		},
	), nil
}

func startLoadingProjectCommits(ctx context.Context, projectID string) error {
	return model.LockProjectE(ctx, projectID, func(project *model.Project) error {
		if project.IsLoadingCommits {
			return ErrDuplicate
		}

		project.IsLoadingCommits = true
		project.MustStore(ctx)

		return nil
	})
}

func doLoadProjectCommits(ctx context.Context, projectID string) error {
	return model.MustLockProjectE(ctx, projectID, func(project *model.Project) error {
		return project.Update(ctx)
	})
}

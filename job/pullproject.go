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

// PullProject pulls a repository from origin.
func PullProject(ctx context.Context, projectID string, priority model.JobPriority) (string, error) {
	if err := startPullingProject(ctx, projectID); err != nil {
		return "", err
	}

	modelCtx := model.GetContext(ctx)

	return modelCtx.Jobs.Add(
		ctx,
		JobNamePullProject,
		projectID,
		priority,
		func(ctx context.Context) error {
			return doPullProject(ctx, projectID)
		},
	), nil
}

func startPullingProject(ctx context.Context, projectID string) error {
	return model.LockProjectE(ctx, projectID, func(project *model.Project) error {
		if project.IsPulling {
			return ErrDuplicate
		}

		project.IsPulling = true
		project.MustStore(ctx)

		return nil
	})
}

func doPullProject(ctx context.Context, projectID string) error {
	return model.MustLockProjectE(ctx, projectID, func(project *model.Project) error {
		return project.Pull(ctx)
	})
}

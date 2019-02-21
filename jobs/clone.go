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

// Clone clones a remote repository locally.
func Clone(ctx context.Context, projectID string, priority models.JobPriority) (string, error) {
	if err := startCloning(ctx, projectID); err != nil {
		return "", err
	}

	modelCtx := models.GetModelContext(ctx)

	return modelCtx.Jobs.Add(
		ctx,
		CloneJob,
		projectID,
		priority,
		func(ctx context.Context) error {
			return doClone(ctx, projectID)
		},
	), nil
}

func startCloning(ctx context.Context, projectID string) error {
	return models.LockProjectE(ctx, projectID, func(project models.Project) error {
		if project.IsCloning {
			return ErrDuplicate
		}

		project.IsCloning = true
		project.MustStore(ctx)

		return nil
	})
}

func doClone(ctx context.Context, projectID string) error {
	return models.MustLockProjectE(ctx, projectID, func(project models.Project) error {
		return project.Clone(ctx)
	})
}

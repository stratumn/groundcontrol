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

// RunTask runs a task.
func RunTask(ctx context.Context, taskID string, env []string, priority model.JobPriority) (string, error) {
	if err := startRunningTask(ctx, taskID); err != nil {
		return "", err
	}

	modelCtx := model.GetContext(ctx)

	return modelCtx.Jobs.Add(
		ctx,
		JobNameRunTask,
		model.MustLoadTask(ctx, taskID).WorkspaceID,
		priority,
		func(ctx context.Context) error {
			return doRunTask(ctx, taskID, env)
		},
	), nil
}

func startRunningTask(ctx context.Context, taskID string) error {
	return model.LockTaskE(ctx, taskID, func(task *model.Task) error {
		if task.Status != model.TaskStatusStopped && task.Status != model.TaskStatusFailed {
			return ErrDuplicate
		}

		task.Status = model.TaskStatusQueued
		task.MustStore(ctx)

		return nil
	})
}

func doRunTask(ctx context.Context, taskID string, env []string) error {
	return model.MustLockTaskE(ctx, taskID, func(task *model.Task) error {
		return task.Run(ctx, env)
	})
}

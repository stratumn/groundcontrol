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
	"io"
	"os/exec"
	"strings"

	"groundcontrol/model"
)

// Run runs a task.
func Run(ctx context.Context, taskID string, env []string, priority model.JobPriority) (string, error) {
	modelCtx := model.GetContext(ctx)
	workspaceID := ""

	err := model.LockTaskE(ctx, taskID, func(task *model.Task) error {
		if task.IsRunning {
			return ErrDuplicate
		}

		workspaceID = task.WorkspaceID
		task.IsRunning = true
		task.MustStore(ctx)

		return nil
	})
	if err != nil {
		return "", err
	}

	return modelCtx.Jobs.Add(
		ctx,
		JobNameRun,
		workspaceID,
		priority,
		func(ctx context.Context) error {
			return doRun(ctx, taskID, env, workspaceID, modelCtx.SystemID)
		},
	), nil
}

func doRun(ctx context.Context, taskID string, env []string, workspaceID string, systemID string) error {
	modelCtx := model.GetContext(ctx)
	log := modelCtx.Log
	pm := modelCtx.PM

	defer func() {
		model.MustLockTask(ctx, taskID, func(task *model.Task) {
			task.IsRunning = false
			task.MustStore(ctx)
		})
	}()

	workspace := model.MustLoadWorkspace(ctx, workspaceID)
	task := model.MustLoadTask(ctx, taskID)
	processGroupID := ""

	for _, stepID := range task.StepsIDs {
		step := model.MustLoadStep(ctx, stepID)

		for _, commandID := range step.CommandsIDs {
			command := model.MustLoadCommand(ctx, commandID)

			for _, projectID := range step.ProjectsIDs {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				project := model.MustLoadProject(ctx, projectID)

				log.InfoWithOwner(ctx, project.ID, command.Command)

				projectPath := modelCtx.GetProjectPath(workspace.Slug, project.Slug)
				parts := strings.Split(command.Command, " ")

				if len(parts) > 0 && parts[0] == "spawn" {
					if processGroupID == "" {
						processGroupID = pm.CreateGroup(ctx, taskID)
					}

					rest := strings.Join(parts[1:], " ")
					pm.Run(ctx, rest, env, processGroupID, project.ID)

					continue
				}

				stdout := model.CreateLineWriter(ctx, log.InfoWithOwner, project.ID)
				stderr := model.CreateLineWriter(ctx, log.WarningWithOwner, project.ID)
				err := run(ctx, command.Command, projectPath, env, stdout, stderr)

				stdout.Close()
				stderr.Close()

				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func run(
	ctx context.Context,
	command string,
	dir string,
	env []string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	cmd := exec.CommandContext(ctx, "bash", "-l", "-c", command)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd.Run()
}

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
	"io"
	"os/exec"
	"strings"

	"github.com/stratumn/groundcontrol/models"
)

// Run runs a task.
func Run(ctx context.Context, taskID string, priority models.JobPriority) (string, error) {
	var (
		taskError   error
		workspaceID string
	)

	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	subs := modelCtx.Subs

	err := nodes.LockTask(taskID, func(task models.Task) {
		if task.IsRunning {
			taskError = ErrDuplicate
			return
		}

		workspaceID = task.WorkspaceID
		task.IsRunning = true
		nodes.MustStoreTask(task)
	})
	if err != nil {
		return "", err
	}
	if taskError != nil {
		return "", taskError
	}

	subs.Publish(models.TaskUpdated, taskID)
	subs.Publish(models.WorkspaceUpdated, workspaceID)

	jobID := modelCtx.Jobs.Add(
		models.GetModelContext(ctx),
		RunJob,
		workspaceID,
		priority,
		func(ctx context.Context) error {
			return doRun(ctx, taskID, workspaceID, modelCtx.SystemID)
		},
	)

	return jobID, nil
}

func doRun(ctx context.Context, taskID string, workspaceID string, systemID string) error {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	subs := modelCtx.Subs
	log := modelCtx.Log
	pm := modelCtx.PM

	defer func() {
		nodes.MustLockTask(taskID, func(task models.Task) {
			task.IsRunning = false
			nodes.MustStoreTask(task)
		})

		subs.Publish(models.TaskUpdated, taskID)
		subs.Publish(models.WorkspaceUpdated, workspaceID)
	}()

	workspace := nodes.MustLoadWorkspace(workspaceID)
	task := nodes.MustLoadTask(taskID)
	processGroupID := ""

	for _, stepID := range task.StepIDs {
		step := nodes.MustLoadStep(stepID)

		for _, commandID := range step.CommandIDs {
			command := nodes.MustLoadCommand(commandID)

			for _, projectID := range step.ProjectIDs {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				project := nodes.MustLoadProject(projectID)

				log.InfoWithOwner(project.ID, command.Command)

				projectPath := modelCtx.GetProjectPath(workspace.Slug, project.Repository, project.Branch)
				parts := strings.Split(command.Command, " ")

				if len(parts) > 0 && parts[0] == "spawn" {
					if processGroupID == "" {
						processGroupID = pm.CreateGroup(ctx, taskID)
					}

					rest := strings.Join(parts[1:], " ")
					pm.Run(ctx, rest, processGroupID, project.ID)

					continue
				}

				stdout := models.CreateLineWriter(log.InfoWithOwner, project.ID)
				stderr := models.CreateLineWriter(log.WarningWithOwner, project.ID)
				err := run(ctx, command.Command, projectPath, stdout, stderr)

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
	stdout io.Writer,
	stderr io.Writer,
) error {
	cmd := exec.CommandContext(ctx, "bash", "-l", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd.Run()
}

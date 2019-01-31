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
	"github.com/stratumn/groundcontrol/pubsub"
)

// Run runs a task.
func Run(
	nodes *models.NodeManager,
	log *models.Logger,
	jobs *models.JobManager,
	pm *models.ProcessManager,
	subs *pubsub.PubSub,
	getProjectPath models.ProjectPathGetter,
	taskID string,
	systemID string,
	priority models.JobPriority,
) (string, error) {
	var (
		taskError   error
		workspaceID string
	)

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

	jobID := jobs.Add(
		RunJob,
		workspaceID,
		priority,
		func(ctx context.Context) error {
			return doRun(
				ctx,
				nodes,
				log,
				pm,
				subs,
				getProjectPath,
				taskID,
				workspaceID,
				systemID,
			)
		},
	)

	return jobID, nil
}

func doRun(
	ctx context.Context,
	nodes *models.NodeManager,
	log *models.Logger,
	pm *models.ProcessManager,
	subs *pubsub.PubSub,
	getProjectPath models.ProjectPathGetter,
	taskID string,
	workspaceID string,
	systemID string,
) error {
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

	for _, step := range task.Steps(nodes) {
		for _, command := range step.Commands {
			for _, project := range step.Projects(nodes) {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				log.InfoWithOwner(project.ID, command)

				projectPath := getProjectPath(workspace.Slug, project.Repository, project.Branch)
				parts := strings.Split(command, " ")

				if len(parts) > 0 && parts[0] == "spawn" {
					if processGroupID == "" {
						processGroupID = pm.CreateGroup(taskID)
					}

					rest := strings.Join(parts[1:], " ")
					pm.Run(rest, processGroupID, project.ID)

					continue
				}

				stdout := models.CreateLineWriter(log.InfoWithOwner, project.ID)
				stderr := models.CreateLineWriter(log.WarningWithOwner, project.ID)
				err := run(ctx, command, projectPath, stdout, stderr)

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

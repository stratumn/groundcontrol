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
	"bufio"
	"io"
	"os/exec"
	"time"

	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/pubsub"
)

// Remember to call close().
func createCommandWriter(
	write func(string, interface{}) string,
	meta interface{},
) io.WriteCloser {
	r, w := io.Pipe()
	scanner := bufio.NewScanner(r)

	go func() {
		for scanner.Scan() {
			write(scanner.Text(), meta)

			// Don't kill the poor browser.
			time.Sleep(10 * time.Millisecond)
		}
	}()

	return w
}

// Run runs a remote repository locally.
func Run(
	nodes *models.NodeManager,
	log *models.Logger,
	jobs *models.JobManager,
	subs *pubsub.PubSub,
	getProjectPath models.ProjectPathGetter,
	taskID string,
) (string, error) {
	var (
		err         error
		workspaceID string
	)

	err = nodes.LockTask(taskID, func(task models.Task) {
		if task.IsRunning {
			err = ErrDuplicate
			return
		}

		workspaceID = task.WorkspaceID
		task.IsRunning = true
		nodes.MustStoreTask(task)
	})
	if err != nil {
		return "", err
	}

	subs.Publish(models.TaskUpdated, taskID)
	subs.Publish(models.WorkspaceUpdated, workspaceID)

	jobID := jobs.Add(RunJob, workspaceID, func() error {
		return doRun(
			nodes,
			log,
			subs,
			getProjectPath,
			taskID,
			workspaceID,
		)
	})

	return jobID, nil
}

func doRun(
	nodes *models.NodeManager,
	log *models.Logger,
	subs *pubsub.PubSub,
	getProjectPath models.ProjectPathGetter,
	taskID string,
	workspaceID string,
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

	for _, step := range task.Steps(nodes) {
		for _, project := range step.Projects(nodes) {
			for _, command := range step.Commands {
				projectPath := getProjectPath(workspace.Slug, project.Repository, project.Branch)

				meta := struct {
					TaskID      string
					WorkspaceID string
					ProjectID   string
					ProjectPath string
					Command     string
				}{
					taskID,
					workspaceID,
					project.ID,
					projectPath,
					command,
				}

				stdout := createCommandWriter(log.Info, meta)
				stderr := createCommandWriter(log.Warning, meta)

				cmd := exec.Command("sh", "-c", command)
				cmd.Dir = projectPath
				cmd.Stdout = stdout
				cmd.Stderr = stderr

				err := cmd.Run()
				stdout.Close()
				stderr.Close()
				return err
			}
		}
	}

	return nil
}

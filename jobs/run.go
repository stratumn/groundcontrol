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
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync/atomic"
	"time"

	"github.com/stratumn/groundcontrol/date"
	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/relay"
)

var (
	nextProcessGroupID uint64
	nextProcessID      uint64
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
	systemID string,
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
			systemID,
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
				projectPath := getProjectPath(workspace.Slug, project.Repository, project.Branch)

				meta := struct {
					TaskID      string
					ProjectID   string
					ProjectPath string
					Command     string
				}{
					taskID,
					project.ID,
					projectPath,
					command,
				}

				parts := strings.Split(command, " ")

				if len(parts) > 0 && parts[0] == "spawn" {
					if processGroupID == "" {
						processGroupID = relay.EncodeID(
							models.NodeTypeProcessGroup,
							fmt.Sprint(atomic.AddUint64(&nextProcessGroupID, 1)),
						)
						nodes.MustStoreProcessGroup(models.ProcessGroup{
							ID:        processGroupID,
							CreatedAt: date.NowFormatted(),
							TaskID:    taskID,
						})
						nodes.MustLockSystem(systemID, func(system models.System) {
							system.ProcessGroupIDs = append(system.ProcessGroupIDs, processGroupID)
							nodes.MustStoreSystem(system)
						})
					}

					rest := strings.Join(parts[1:], " ")
					spawn(nodes, log, subs, rest, projectPath, processGroupID)
					return nil
				}

				stdout := createCommandWriter(log.Info, meta)
				stderr := createCommandWriter(log.Warning, meta)
				err := run(command, projectPath, stdout, stderr)

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

func spawn(
	nodes *models.NodeManager,
	log *models.Logger,
	subs *pubsub.PubSub,
	command string,
	projectPath string,
	processGroupID string,
) {
	meta := struct {
		ProjectPath    string
		Command        string
		ProcessGroupID string
	}{
		projectPath,
		command,
		processGroupID,
	}

	id := relay.EncodeID(
		models.NodeTypeProcess,
		fmt.Sprint(atomic.AddUint64(&nextProcessID, 1)),
	)

	process := models.Process{
		ID:             id,
		Command:        command,
		Status:         models.ProcessStatusRunning,
		ProcessGroupID: processGroupID,
	}

	nodes.MustStoreProcess(process)

	nodes.MustLockProcessGroup(processGroupID, func(processGroup models.ProcessGroup) {
		processGroup.ProcessIDs = append(processGroup.ProcessIDs, id)
		nodes.MustStoreProcessGroup(processGroup)
	})

	subs.Publish(models.ProcessUpserted, id)
	subs.Publish(models.ProcessGroupUpserted, processGroupID)

	go func() {
		stdout := createCommandWriter(log.Info, meta)
		stderr := createCommandWriter(log.Warning, meta)

		err := run(command, projectPath, stdout, stderr)

		nodes.MustLockProcess(id, func(process models.Process) {
			if err != nil {
				log.Error("Process Failed", meta)
				process.Status = models.ProcessStatusFailed
			} else {
				log.Info("Process Done", meta)
				process.Status = models.ProcessStatusDone
			}

			nodes.MustStoreProcess(process)
		})

		subs.Publish(models.ProcessUpserted, id)
		subs.Publish(models.ProcessGroupUpserted, processGroupID)
	}()
}

func run(
	command string,
	dir string,
	stdout io.Writer,
	stderr io.Writer,
) error {
	cmd := exec.Command("bash", "-l", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd.Run()
}

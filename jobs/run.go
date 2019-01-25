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
	"time"

	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/pubsub"
)

// Run runs a remote repository locally.
func Run(
	nodes *models.NodeManager,
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
	subs *pubsub.PubSub,
	getProjectPath models.ProjectPathGetter,
	taskID string,
	workspaceID string,
) error {
	//task := nodes.MustLoadTask(taskID)

	defer func() {
		nodes.MustLockTask(taskID, func(task models.Task) {
			task.IsRunning = false
			nodes.MustStoreTask(task)
		})

		subs.Publish(models.TaskUpdated, taskID)
		subs.Publish(models.WorkspaceUpdated, workspaceID)
	}()

	time.Sleep(10 * time.Second)
	return nil
}

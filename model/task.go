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

package model

import (
	"context"
	"io"
	"os/exec"

	"groundcontrol/appcontext"
	"groundcontrol/util"
)

// Run executes the commands in the task.
// Env is the environment of the Task. Each entry is of the form 'key=value'.
func (n *Task) Run(ctx context.Context, env []string) error {
	var err error
	defer func() {
		if err == nil {
			n.Status = TaskStatusStopped
		} else {
			n.Status = TaskStatusFailed
		}
		n.MustStore(ctx)
	}()
	n.Status = TaskStatusRunning
	n.MustStore(ctx)

	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	workspace := n.Workspace(ctx)

	for _, stepID := range n.StepsIDs {
		step := MustLoadStep(ctx, stepID)
		n.CurrentStepID = stepID

		for _, projectID := range step.ProjectsIDs {
			project := MustLoadProject(ctx, projectID)
			n.CurrentProjectID = projectID

			for _, commandID := range step.CommandsIDs {
				command := MustLoadCommand(ctx, commandID)
				n.CurrentCommandID = commandID
				n.MustStore(ctx)

				select {
				case <-ctx.Done():
					err = ctx.Err()
					return err
				default:
				}

				log.InfoWithOwner(ctx, project.ID, command.Command)

				projectPath := appCtx.GetProjectPath(workspace.Slug, project.Slug)
				stdout := util.LineSplitter(ctx, log.InfoWithOwner, project.ID)
				stderr := util.LineSplitter(ctx, log.WarningWithOwner, project.ID)
				err = n.runCmd(ctx, command.Command, projectPath, env, stdout, stderr)

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

func (n *Task) runCmd(
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

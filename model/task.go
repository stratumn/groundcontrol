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
)

// Run executes the commands in the task.
// Env is the environment of the Task. Each entry is of the form 'key=value'.
func (n *Task) Run(ctx context.Context, env []string) error {
	defer func() {
		n.IsRunning = false
		n.MustStore(ctx)
	}()

	n.IsRunning = true
	n.MustStore(ctx)

	modelCtx := GetContext(ctx)
	log := modelCtx.Log
	workspace := n.Workspace(ctx)

	for _, stepID := range n.StepsIDs {
		step := MustLoadStep(ctx, stepID)

		for _, commandID := range step.CommandsIDs {
			command := MustLoadCommand(ctx, commandID)

			for _, projectID := range step.ProjectsIDs {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				project := MustLoadProject(ctx, projectID)

				log.InfoWithOwner(ctx, project.ID, command.Command)

				projectPath := modelCtx.GetProjectPath(workspace.Slug, project.Slug)
				stdout := CreateLineWriter(ctx, log.InfoWithOwner, project.ID)
				stderr := CreateLineWriter(ctx, log.WarningWithOwner, project.ID)
				err := n.runCmd(ctx, command.Command, projectPath, env, stdout, stderr)

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

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

	for _, stepID := range n.StepsIDs {
		step := MustLoadStep(ctx, stepID)
		n.CurrentStepID = stepID
		if err = n.runStep(ctx, step, env); err != nil {
			return err
		}
	}
	return nil
}

func (n *Task) runStep(ctx context.Context, step *Step, env []string) error {
	for _, projectID := range step.ProjectsIDs {
		project := MustLoadProject(ctx, projectID)
		n.CurrentProjectID = projectID
		if err := n.runStepOnProject(ctx, step, project, env); err != nil {
			return err
		}
	}
	return nil
}

func (n *Task) runStepOnProject(ctx context.Context, step *Step, project *Project, env []string) error {
	for _, commandID := range step.CommandsIDs {
		command := MustLoadCommand(ctx, commandID)
		n.CurrentCommandID = commandID
		n.MustStore(ctx)
		if err := n.runCommand(ctx, project, command, env); err != nil {
			return err
		}
	}
	return nil
}

func (n *Task) runCommand(ctx context.Context, project *Project, command *Command, env []string) error {
	log := appcontext.Get(ctx).Log
	log.InfoWithOwner(ctx, project.ID, command.Command)
	stdout := util.LineSplitter(ctx, log.InfoWithOwner, project.ID)
	stderr := util.LineSplitter(ctx, log.WarningWithOwner, project.ID)
	cmd := exec.CommandContext(ctx, "bash", "-l", "-c", command.Command)
	cmd.Dir = project.Path(ctx)
	cmd.Env = env
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	stdout.Close()
	stderr.Close()
	return err
}

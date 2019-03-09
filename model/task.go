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
	"os"

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
	if len(step.ProjectsIDs) < 1 {
		n.CurrentProjectID = ""
		return n.runStepCmds(ctx, step, "", env)
	}
	for _, projectID := range step.ProjectsIDs {
		project := MustLoadProject(ctx, projectID)
		if err := project.EnsureCloned(ctx); err != nil {
			return err
		}
		n.CurrentProjectID = projectID
		if err := n.runStepCmds(ctx, step, project.Path(ctx), env); err != nil {
			return err
		}
	}
	return nil
}

func (n *Task) runStepCmds(ctx context.Context, step *Step, dir string, env []string) error {
	runner, close, err := n.createRunner(ctx, dir, env)
	if err != nil {
		return err
	}
	defer close()
	log := appcontext.Get(ctx).Log
	for _, commandID := range step.CommandsIDs {
		command := MustLoadCommand(ctx, commandID)
		n.CurrentCommandID = commandID
		n.MustStore(ctx)
		log.InfoWithOwner(ctx, n.ID, command.Command)
		if err := runner.Run(ctx, command.Command); err != nil {
			return err
		}
	}
	return nil
}

func (n *Task) createRunner(ctx context.Context, dir string, env []string) (appcontext.Runner, func(), error) {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	stdout := util.LineSplitter(ctx, log.InfoWithOwner, n.ID)
	stderr := util.LineSplitter(ctx, log.WarningWithOwner, n.ID)
	close := func() {
		stdout.Close()
		stderr.Close()
	}
	env = append(os.Environ(), env...)
	runner, err := appCtx.NewRunner(stdout, stderr, dir, env)
	if err != nil {
		close()
		return nil, nil, err
	}
	return runner, close, nil
}

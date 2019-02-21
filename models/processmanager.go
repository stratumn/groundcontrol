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

package models

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"groundcontrol/relay"
)

// ProcessManager manages creating and running jobs.
type ProcessManager struct {
	commands sync.Map
	lastID   uint64

	runningCounter int64
	doneCounter    int64
	failedCounter  int64
}

// NewProcessManager creates a ProcessManager.
func NewProcessManager() *ProcessManager {
	return &ProcessManager{}
}

// CreateGroup creates a new ProcessGroup and returns its ID.
func (p *ProcessManager) CreateGroup(ctx context.Context, taskID string) string {
	modelCtx := GetModelContext(ctx)

	id := relay.EncodeID(
		NodeTypeProcessGroup,
		fmt.Sprint(atomic.AddUint64(&p.lastID, 1)),
	)

	group := ProcessGroup{
		ID:        id,
		CreatedAt: DateTime(time.Now()),
		TaskID:    taskID,
	}

	group.MustStore(ctx)

	MustLockSystem(ctx, modelCtx.SystemID, func(system System) {
		system.ProcessGroupIDs = append(
			[]string{id},
			system.ProcessGroupIDs...,
		)

		system.MustStore(ctx)
	})

	return id
}

// Run launches a new Process and adds it to a ProcessGroup.
func (p *ProcessManager) Run(
	ctx context.Context,
	command string,
	env []string,
	processGroupID string,
	projectID string,
) string {
	id := relay.EncodeID(
		NodeTypeProcess,
		fmt.Sprint(atomic.AddUint64(&p.lastID, 1)),
	)

	process := Process{
		ID:             id,
		Command:        command,
		Env:            env,
		ProcessGroupID: processGroupID,
		ProjectID:      projectID,
	}

	process.MustStore(ctx)
	MustLockProcessGroup(ctx, processGroupID, func(processGroup ProcessGroup) {
		processGroup.ProcessIDs = append([]string{id}, processGroup.ProcessIDs...)
		processGroup.MustStore(ctx)
	})

	p.exec(ctx, id)

	return id
}

// Start starts a process that was stopped.
func (p *ProcessManager) Start(ctx context.Context, processID string) error {
	err := LockProcessE(ctx, processID, func(process Process) error {
		switch process.Status {
		case ProcessStatusRunning, ProcessStatusStopping:
			return ErrNotStopped
		case ProcessStatusDone:
			atomic.AddInt64(&p.doneCounter, -1)
		case ProcessStatusFailed:
			atomic.AddInt64(&p.failedCounter, -1)
		}
		return nil
	})
	if err != nil {
		return err
	}

	p.exec(ctx, processID) // will publish metrics

	return nil
}

// Stop stops a running process.
func (p *ProcessManager) Stop(ctx context.Context, processID string) error {
	return LockProcessE(ctx, processID, func(process Process) error {
		if process.Status != ProcessStatusRunning {
			return ErrNotRunning
		}

		process.Status = ProcessStatusStopping
		process.MustStore(ctx)

		actual, ok := p.commands.Load(processID)
		if !ok {
			panic("command not found")
		}
		cmd := actual.(*exec.Cmd)

		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err != nil {
			return err
		}

		return syscall.Kill(-pgid, syscall.SIGINT)
	})
}

// Clean terminates all running processes.
func (p *ProcessManager) Clean(ctx context.Context) {
	modelCtx := GetModelContext(ctx)
	waitGroup := sync.WaitGroup{}

	p.commands.Range(func(k, _ interface{}) bool {
		processID := k.(string)

		modelCtx.Log.DebugWithOwner(ctx, processID, "stopping process")

		if err := p.Stop(ctx, processID); err != nil {
			modelCtx.Log.ErrorWithOwner(
				ctx,
				processID,
				"failed to stop process because %s",
				err.Error(),
			)
			return true
		}

		waitGroup.Add(1)

		processCtx, cancel := context.WithCancel(ctx)

		go func() {
			<-processCtx.Done()
			waitGroup.Done()
		}()

		modelCtx.Subs.Subscribe(processCtx, MessageTypeProcessStored, 0, func(msg interface{}) {
			id := msg.(string)
			if id != processID {
				return
			}

			process := MustLoadProcess(ctx, id)

			switch process.Status {
			case ProcessStatusDone, ProcessStatusFailed:
				modelCtx.Log.DebugWithOwner(ctx, processID, "process stopped")
				cancel()
			}
		})

		return true
	})

	waitGroup.Wait()
}

func (p *ProcessManager) exec(ctx context.Context, id string) {
	modelCtx := GetModelContext(ctx)

	MustLockProcess(ctx, id, func(process Process) {
		project := process.Project(ctx)
		workspace := project.Workspace(ctx)

		dir := modelCtx.GetProjectPath(workspace.Slug, project.Slug)

		stdout := CreateLineWriter(ctx, modelCtx.Log.InfoWithOwner, project.ID)
		stderr := CreateLineWriter(ctx, modelCtx.Log.WarningWithOwner, project.ID)
		cmd := exec.Command("bash", "-l", "-c", process.Command)
		cmd.Dir = dir
		cmd.Env = process.Env
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		err := cmd.Start()
		if err == nil {
			process.Status = ProcessStatusRunning
			atomic.AddInt64(&p.runningCounter, 1)
		} else {
			process.Status = ProcessStatusFailed
			atomic.AddInt64(&p.failedCounter, 1)
		}

		process.MustStore(ctx)
		p.updateMetrics(ctx)

		if err != nil {
			modelCtx.Log.ErrorWithOwner(
				ctx,
				project.ID,
				"process failed because %s",
				err.Error(),
			)
			stdout.Close()
			stderr.Close()
			return
		}

		modelCtx.Log.DebugWithOwner(ctx, project.ID, "process is running")
		p.commands.Store(id, cmd)

		go func() {
			err := cmd.Wait()

			MustLockProcess(ctx, id, func(process Process) {
				p.commands.Delete(id)

				if err == nil {
					process.Status = ProcessStatusDone
					atomic.AddInt64(&p.doneCounter, 1)
					modelCtx.Log.DebugWithOwner(ctx, project.ID, "process done")
				} else {
					process.Status = ProcessStatusFailed
					atomic.AddInt64(&p.failedCounter, 1)
					modelCtx.Log.ErrorWithOwner(
						ctx,
						project.ID,
						"process failed because %s",
						err.Error(),
					)
				}

				atomic.AddInt64(&p.runningCounter, -1)
				process.MustStore(ctx)
			})

			p.updateMetrics(ctx)

			stdout.Close()
			stderr.Close()
		}()
	})
}

func (p *ProcessManager) updateMetrics(ctx context.Context) {
	modelCtx := GetModelContext(ctx)
	system := MustLoadSystem(ctx, modelCtx.SystemID)

	MustLockProcessMetrics(ctx, system.ProcessMetricsID, func(metrics ProcessMetrics) {
		metrics.Running = int(atomic.LoadInt64(&p.runningCounter))
		metrics.Done = int(atomic.LoadInt64(&p.doneCounter))
		metrics.Failed = int(atomic.LoadInt64(&p.failedCounter))
		metrics.MustStore(ctx)
	})
}

// CreateLineWriter creates a writer with a line splitter.
// Remember to call close().
func CreateLineWriter(
	ctx context.Context,
	write func(
		ctx context.Context,
		ownerID,
		message string,
		a ...interface{},
	) string,
	ownerID string,
	a ...interface{},
) io.WriteCloser {
	r, w := io.Pipe()
	scanner := bufio.NewScanner(r)

	go func() {
		for scanner.Scan() {
			write(ctx, ownerID, scanner.Text(), a...)

			// Don't kill the poor browser.
			time.Sleep(10 * time.Millisecond)
		}
	}()

	return w
}

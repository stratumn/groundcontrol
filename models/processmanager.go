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

	"github.com/stratumn/groundcontrol/relay"
)

// ProcessManager manages creating and running jobs.
type ProcessManager struct {
	commands sync.Map
	nextID   uint64

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
		fmt.Sprint(atomic.AddUint64(&p.nextID, 1)),
	)

	group := ProcessGroup{
		ID:        id,
		CreatedAt: DateTime(time.Now()),
		TaskID:    taskID,
	}

	modelCtx.Nodes.MustStoreProcessGroup(group)
	modelCtx.Subs.Publish(ProcessGroupUpserted, id)

	modelCtx.Nodes.MustLockSystem(modelCtx.SystemID, func(system System) {
		system.ProcessGroupIDs = append(
			[]string{id},
			system.ProcessGroupIDs...,
		)

		modelCtx.Nodes.MustStoreSystem(system)
	})

	return id
}

// Run launches a new Process and adds it to a ProcessGroup.
func (p *ProcessManager) Run(
	ctx context.Context,
	command string,
	processGroupID string,
	projectID string,
) string {
	modelCtx := GetModelContext(ctx)

	id := relay.EncodeID(
		NodeTypeProcess,
		fmt.Sprint(atomic.AddUint64(&p.nextID, 1)),
	)

	process := Process{
		ID:             id,
		Command:        command,
		ProcessGroupID: processGroupID,
		ProjectID:      projectID,
	}

	modelCtx.Nodes.MustStoreProcess(process)
	modelCtx.Nodes.MustLockProcessGroup(processGroupID, func(processGroup ProcessGroup) {
		processGroup.ProcessIDs = append([]string{id}, processGroup.ProcessIDs...)
		modelCtx.Nodes.MustStoreProcessGroup(processGroup)
	})

	p.exec(ctx, id)

	return id
}

// Start starts a process that was stopped.
func (p *ProcessManager) Start(ctx context.Context, processID string) error {
	modelCtx := GetModelContext(ctx)

	err := modelCtx.Nodes.LockProcessE(processID, func(process Process) error {
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
	modelCtx := GetModelContext(ctx)

	return modelCtx.Nodes.LockProcessE(processID, func(process Process) error {
		if process.Status != ProcessStatusRunning {
			return ErrNotRunning
		}

		process.Status = ProcessStatusStopping
		modelCtx.Nodes.MustStoreProcess(process)

		modelCtx.Subs.Publish(ProcessUpserted, processID)
		modelCtx.Subs.Publish(ProcessGroupUpserted, process.ProcessGroupID)

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

		modelCtx.Log.InfoWithOwner(processID, "stopping process")

		if err := p.Stop(ctx, processID); err != nil {
			modelCtx.Log.ErrorWithOwner(processID, "failed to stop process because %s", err.Error())
			return true
		}

		waitGroup.Add(1)

		processCtx, cancel := context.WithCancel(ctx)

		go func() {
			<-processCtx.Done()
			waitGroup.Done()
		}()

		modelCtx.Subs.Subscribe(processCtx, ProcessUpserted, func(msg interface{}) {
			id := msg.(string)
			if id != processID {
				return
			}

			process := modelCtx.Nodes.MustLoadProcess(id)

			switch process.Status {
			case ProcessStatusDone, ProcessStatusFailed:
				modelCtx.Log.InfoWithOwner(processID, "process stopped")
				cancel()
			}
		})

		return true
	})

	waitGroup.Wait()
}

func (p *ProcessManager) exec(ctx context.Context, id string) {
	modelCtx := GetModelContext(ctx)

	modelCtx.Nodes.MustLockProcess(id, func(process Process) {
		project := process.Project(ctx)
		workspace := project.Workspace(ctx)

		dir := modelCtx.GetProjectPath(
			workspace.Slug,
			project.Repository,
			project.Branch,
		)

		stdout := CreateLineWriter(modelCtx.Log.InfoWithOwner, project.ID)
		stderr := CreateLineWriter(modelCtx.Log.WarningWithOwner, project.ID)
		cmd := exec.Command("bash", "-l", "-c", process.Command)
		cmd.Dir = dir
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

		modelCtx.Nodes.MustStoreProcess(process)
		modelCtx.Subs.Publish(ProcessUpserted, id)
		modelCtx.Subs.Publish(ProcessGroupUpserted, process.ProcessGroupID)
		p.publishMetrics(ctx)

		if err != nil {
			modelCtx.Log.ErrorWithOwner(project.ID, "process failed because %s", err.Error())
			stdout.Close()
			stderr.Close()
			return
		}

		modelCtx.Log.InfoWithOwner(project.ID, "process is running")
		p.commands.Store(id, cmd)

		go func() {
			err := cmd.Wait()

			modelCtx.Nodes.MustLockProcess(id, func(process Process) {
				p.commands.Delete(id)

				if err == nil {
					process.Status = ProcessStatusDone
					atomic.AddInt64(&p.doneCounter, 1)
					modelCtx.Log.InfoWithOwner(project.ID, "process done")
				} else {
					process.Status = ProcessStatusFailed
					atomic.AddInt64(&p.failedCounter, 1)
					modelCtx.Log.ErrorWithOwner(project.ID, "process failed because %s", err.Error())
				}

				atomic.AddInt64(&p.runningCounter, -1)
				modelCtx.Nodes.MustStoreProcess(process)
			})

			modelCtx.Subs.Publish(ProcessUpserted, id)
			modelCtx.Subs.Publish(ProcessGroupUpserted, process.ProcessGroupID)
			p.publishMetrics(ctx)

			stdout.Close()
			stderr.Close()
		}()
	})
}

func (p *ProcessManager) publishMetrics(ctx context.Context) {
	modelCtx := GetModelContext(ctx)
	system := modelCtx.Nodes.MustLoadSystem(modelCtx.SystemID)

	modelCtx.Nodes.MustLockProcessMetrics(system.ProcessMetricsID, func(metrics ProcessMetrics) {
		metrics.Running = int(atomic.LoadInt64(&p.runningCounter))
		metrics.Done = int(atomic.LoadInt64(&p.doneCounter))
		metrics.Failed = int(atomic.LoadInt64(&p.failedCounter))
		modelCtx.Nodes.MustStoreProcessMetrics(metrics)
	})

	modelCtx.Subs.Publish(ProcessMetricsUpdated, system.ProcessMetricsID)
}

// CreateLineWriter creates a writer with a line splitter.
// Remember to call close().
func CreateLineWriter(
	write func(ownerID, message string, a ...interface{}) string,
	ownerID string,
	a ...interface{},
) io.WriteCloser {
	r, w := io.Pipe()
	scanner := bufio.NewScanner(r)

	go func() {
		for scanner.Scan() {
			write(ownerID, scanner.Text(), a...)

			// Don't kill the poor browser.
			time.Sleep(10 * time.Millisecond)
		}
	}()

	return w
}

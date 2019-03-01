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

package service

import (
	"context"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"

	"groundcontrol/appcontext"
	"groundcontrol/model"
	"groundcontrol/util"
)

// Manager manages running and stopping services.
type Manager struct {
	commands sync.Map

	stoppedCounter  int64
	startingCounter int64
	runningCounter  int64
	stoppingCounter int64
	failedCounter   int64
}

// NewManager creates a Manager.
func NewManager() *Manager {
	return &Manager{}
}

// Start starts a Service and its dependencies.
func (m *Manager) Start(ctx context.Context, serviceID string, env []string) error {
	service, err := model.LoadService(ctx, serviceID)
	if err != nil {
		return err
	}
	for _, depID := range service.DependenciesIDs {
		if err := m.startService(ctx, depID, env); err != nil {
			return err
		}
	}
	return nil
}

// Stop stops a running Service.
func (m *Manager) Stop(ctx context.Context, serviceID string) error {
	return model.LockServiceE(ctx, serviceID, func(service *model.Service) error {
		if service.Status != model.ServiceStatusRunning {
			return model.ErrNotRunning
		}
		m.setStatus(ctx, service, model.ServiceStatusStopping)
		service.MustStore(ctx)
		actual, _ := m.commands.Load(serviceID)
		cmd := actual.(*exec.Cmd)
		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err != nil {
			return err
		}
		return syscall.Kill(-pgid, syscall.SIGINT)
	})
}

// Clean terminates all running Services.
func (m *Manager) Clean(ctx context.Context) {
	appCtx := appcontext.Get(ctx)
	lastMsgID := appCtx.Subs.LastMessageID()
	waitGroup := sync.WaitGroup{}
	m.commands.Range(func(k, _ interface{}) bool {
		serviceID, ok := k.(string)
		if !ok {
			return true
		}
		log := appCtx.Log
		log.DebugWithOwner(ctx, serviceID, "stopping service")
		if err := m.Stop(ctx, serviceID); err != nil {
			log.ErrorWithOwner(ctx, serviceID, "failed to stop service because %s", err.Error())
			return true
		}
		m.waitTillDone(ctx, serviceID, lastMsgID)
		return true
	})
	waitGroup.Wait()
}

func (m *Manager) startService(ctx context.Context, serviceID string, env []string) error {
	return model.LockServiceE(ctx, serviceID, func(service *model.Service) error {
		switch service.Status {
		case model.ServiceStatusStarting, model.ServiceStatusStopping:
			return ErrNotStopped
		case model.ServiceStatusRunning:
			return nil
		}
		m.setStatus(ctx, service, model.ServiceStatusStarting)
		service.MustStore(ctx)
		stdout, stderr, clean := m.buildWriters(ctx, service)
		fail := func() {
			clean()
			m.setStatus(ctx, service, model.ServiceStatusFailed)
			service.MustStore(ctx)
		}
		cmd, err := m.buildCmd(ctx, service, env)
		if err != nil {
			fail()
			return err
		}
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := m.startCmd(ctx, cmd, service, clean); err != nil {
			fail()
		}
		return err
	})
}

func (m *Manager) buildWriters(ctx context.Context, service *model.Service) (io.WriteCloser, io.WriteCloser, func()) {
	log := appcontext.Get(ctx).Log
	ownerID := service.OwnerID()
	stdout := util.LineSplitter(ctx, log.InfoWithOwner, ownerID)
	stderr := util.LineSplitter(ctx, log.WarningWithOwner, ownerID)
	clean := func() {
		stdout.Close()
		stderr.Close()
	}
	return stdout, stderr, clean
}

func (m *Manager) buildCmd(ctx context.Context, service *model.Service, env []string) (*exec.Cmd, error) {
	cmd := exec.Command("bash", "-l", "-c", service.Command)
	cmd.Env = env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if project := service.Project(ctx); project != nil {
		if err := project.EnsureCloned(ctx); err != nil {
			return nil, err
		}
		appCtx := appcontext.Get(ctx)
		workspace := project.Workspace(ctx)
		cmd.Dir = appCtx.GetProjectPath(workspace.Slug, project.Slug)
	}
	return cmd, nil
}

func (m *Manager) startCmd(ctx context.Context, cmd *exec.Cmd, service *model.Service, clean func()) error {
	if err := m.runBeforeTasks(ctx, service, cmd.Env); err != nil {
		return err
	}
	log := appcontext.Get(ctx).Log
	log.InfoWithOwner(ctx, service.OwnerID(), service.Command)
	if err := cmd.Start(); err != nil {
		return err
	}
	m.setStatus(ctx, service, model.ServiceStatusRunning)
	service.MustStore(ctx)
	m.commands.Store(service.ID, cmd)
	go m.runCmd(ctx, cmd, service.ID, clean)
	return nil
}

func (m *Manager) runCmd(ctx context.Context, cmd *exec.Cmd, serviceID string, clean func()) {
	err := cmd.Wait()
	clean()
	model.MustLockService(ctx, serviceID, func(service *model.Service) {
		m.commands.Delete(serviceID)
		taskErr := m.runAfterTasks(ctx, service, cmd.Env)
		// Prioritize the command error over the task error.
		if err != nil && taskErr != nil {
			err = taskErr
		}
		log := appcontext.Get(ctx).Log
		ownerID := service.OwnerID()
		if err == nil {
			m.setStatus(ctx, service, model.ServiceStatusStopped)
			log.DebugWithOwner(ctx, ownerID, "service done")
		} else {
			m.setStatus(ctx, service, model.ServiceStatusFailed)
			log.ErrorWithOwner(ctx, ownerID, "service failed because %s", err.Error())
		}
		service.MustStore(ctx)
	})
}

func (m *Manager) runBeforeTasks(ctx context.Context, service *model.Service, env []string) error {
	return m.runTasks(ctx, service.BeforeIDs, env)
}

func (m *Manager) runAfterTasks(ctx context.Context, service *model.Service, env []string) error {
	return m.runTasks(ctx, service.AfterIDs, env)
}

func (m *Manager) runTasks(ctx context.Context, taskIDs []string, env []string) error {
	for _, taskID := range taskIDs {
		err := model.MustLockTaskE(ctx, taskID, func(task *model.Task) error {
			return task.Run(ctx, env)
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// waitTillDone blocks until the Service exists.
func (m *Manager) waitTillDone(ctx context.Context, serviceID string, lastMsgID uint64) {
	subsCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	subs := appcontext.Get(ctx).Subs
	subs.Subscribe(subsCtx, model.MessageTypeServiceStored, lastMsgID, func(msg interface{}) {
		id := msg.(string)
		if id != serviceID {
			return
		}
		service := model.MustLoadService(ctx, id)
		switch service.Status {
		case model.ServiceStatusStopped, model.ServiceStatusFailed:
			cancel()
		}
	})
	<-subsCtx.Done()
}

func (m *Manager) setStatus(ctx context.Context, service *model.Service, status model.ServiceStatus) {
	was := service.Status
	if was == status {
		return
	}
	service.Status = status
	m.decCounter(was)
	m.incCounter(status)
	m.storeMetrics(ctx)
}

func (m *Manager) incCounter(status model.ServiceStatus) {
	m.addToCounter(status, 1)
}

func (m *Manager) decCounter(status model.ServiceStatus) {
	m.addToCounter(status, -1)
}

func (m *Manager) addToCounter(status model.ServiceStatus, delta int64) {
	switch status {
	case model.ServiceStatusStopped:
		atomic.AddInt64(&m.stoppedCounter, delta)
	case model.ServiceStatusStarting:
		atomic.AddInt64(&m.startingCounter, delta)
	case model.ServiceStatusRunning:
		atomic.AddInt64(&m.runningCounter, delta)
	case model.ServiceStatusStopping:
		atomic.AddInt64(&m.stoppingCounter, delta)
	case model.ServiceStatusFailed:
		atomic.AddInt64(&m.failedCounter, delta)
	}
}

func (m *Manager) storeMetrics(ctx context.Context) {
	appCtx := appcontext.Get(ctx)
	system := model.MustLoadSystem(ctx, appCtx.SystemID)
	model.MustLockServiceMetrics(ctx, system.ServiceMetricsID, func(metrics *model.ServiceMetrics) {
		metrics.Stopped = int(atomic.LoadInt64(&m.stoppedCounter))
		metrics.Starting = int(atomic.LoadInt64(&m.startingCounter))
		metrics.Running = int(atomic.LoadInt64(&m.runningCounter))
		metrics.Stopping = int(atomic.LoadInt64(&m.stoppingCounter))
		metrics.Failed = int(atomic.LoadInt64(&m.failedCounter))
		metrics.MustStore(ctx)
	})
}

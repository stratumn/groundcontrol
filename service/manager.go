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
	"os"
	"sync"
	"sync/atomic"

	"groundcontrol/appcontext"
	"groundcontrol/model"
	"groundcontrol/util"
)

// Manager manages running and stopping services.
type Manager struct {
	cancels sync.Map

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

// Stop stops a running Service. If the service isn't running it returns ErrStatus.
func (m *Manager) Stop(ctx context.Context, serviceID string) error {
	return model.LockServiceE(ctx, serviceID, func(service *model.Service) error {
		if service.Status != model.ServiceStatusRunning {
			return ErrStatus
		}
		m.setStatus(ctx, service, model.ServiceStatusStopping)
		service.MustStore(ctx)
		actual, _ := m.cancels.Load(serviceID)
		actual.(context.CancelFunc)()
		return nil
	})
}

// Clean terminates all running Services.
func (m *Manager) Clean(ctx context.Context) {
	appCtx := appcontext.Get(ctx)
	lastMsgID := appCtx.Subs.LastMessageID()
	waitGroup := sync.WaitGroup{}
	m.cancels.Range(func(k, _ interface{}) bool {
		serviceID, ok := k.(string)
		if !ok {
			return true
		}
		log := appCtx.Log
		log.DebugWithOwner(ctx, appCtx.SystemID, "stopping service")
		if err := m.Stop(ctx, serviceID); err != nil {
			log.ErrorWithOwner(ctx, appCtx.SystemID, "failed to stop service because %s", err.Error())
			return true
		}
		waitGroup.Add(1)
		go func() {
			m.waitTillDone(ctx, serviceID, lastMsgID)
			waitGroup.Done()
		}()
		return true
	})
	waitGroup.Wait()
}

func (m *Manager) startService(ctx context.Context, serviceID string, env []string) error {
	return model.LockServiceE(ctx, serviceID, func(service *model.Service) error {
		switch service.Status {
		case model.ServiceStatusStarting, model.ServiceStatusStopping:
			return ErrStatus
		case model.ServiceStatusRunning:
			return nil
		}
		m.setStatus(ctx, service, model.ServiceStatusStarting)
		service.MustStore(ctx)
		fail := func() {
			m.setStatus(ctx, service, model.ServiceStatusFailed)
			service.MustStore(ctx)
		}
		runner, close, err := m.createRunner(ctx, service, env)
		if err != nil {
			fail()
			return err
		}
		if err := m.launchService(ctx, runner, service, env, close); err != nil {
			fail()
			close()
		}
		return err
	})
}

func (m *Manager) createWriters(ctx context.Context, service *model.Service) (io.WriteCloser, io.WriteCloser, func()) {
	log := appcontext.Get(ctx).Log
	stdout := util.LineSplitter(ctx, log.InfoWithOwner, service.ID)
	stderr := util.LineSplitter(ctx, log.WarningWithOwner, service.ID)
	close := func() {
		stdout.Close()
		stderr.Close()
	}
	return stdout, stderr, close
}

func (m *Manager) createRunner(ctx context.Context, service *model.Service, env []string) (appcontext.Runner, func(), error) {
	appCtx := appcontext.Get(ctx)
	stdout, stderr, close := m.createWriters(ctx, service)
	dir := ""
	if project := service.Project(ctx); project != nil {
		if err := project.EnsureCloned(ctx); err != nil {
			close()
			return nil, nil, err
		}
		workspace := project.Workspace(ctx)
		dir = appCtx.GetProjectPath(workspace.Slug, project.Slug)
	}
	env = append(os.Environ(), env...)
	runner, err := appCtx.NewRunner(stdout, stderr, dir, env, appCtx.RunnerGracefulShutdownTimeout)
	if err != nil {
		close()
		return nil, nil, err
	}
	return runner, close, nil
}

func (m *Manager) launchService(ctx context.Context, runner appcontext.Runner, service *model.Service, env []string, close func()) error {
	if err := m.runBeforeTasks(ctx, service, env); err != nil {
		return err
	}
	m.setStatus(ctx, service, model.ServiceStatusRunning)
	service.MustStore(ctx)
	runCtx, cancel := m.createCtx(ctx)
	m.cancels.Store(service.ID, cancel)
	go m.runService(runCtx, runner, service, env, close)
	return nil
}

func (m *Manager) runService(ctx context.Context, runner appcontext.Runner, service *model.Service, env []string, close func()) {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	log.InfoWithOwner(ctx, service.ID, service.Command)
	err := runner.Run(ctx, service.Command)
	close()
	model.MustLockService(ctx, service.ID, func(service *model.Service) {
		m.cancels.Delete(service.ID)
		taskErr := m.runAfterTasks(ctx, service, env)
		// Prioritize the command error over the task error.
		if err != nil && taskErr != nil {
			err = taskErr
		}
		if err == nil {
			m.setStatus(ctx, service, model.ServiceStatusStopped)
			log.DebugWithOwner(ctx, appCtx.SystemID, "service done")
		} else {
			m.setStatus(ctx, service, model.ServiceStatusFailed)
			log.ErrorWithOwner(ctx, appCtx.SystemID, "service failed because %s", err.Error())
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

// waitTillDone blocks until the Service exits.
func (m *Manager) waitTillDone(ctx context.Context, serviceID string, lastMsgID uint64) {
	subsCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	subs := appcontext.Get(ctx).Subs
	subs.Subscribe(subsCtx, model.MessageTypeServiceStored, lastMsgID, func(msg interface{}) {
		service := msg.(*model.Service)
		if service.ID != serviceID {
			return
		}
		switch service.Status {
		case model.ServiceStatusStopped, model.ServiceStatusFailed:
			cancel()
		}
	})
	<-subsCtx.Done()
}

func (m *Manager) createCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	appCtx := appcontext.Get(ctx)
	ctx = appcontext.With(context.Background(), appCtx)
	return context.WithCancel(ctx)
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

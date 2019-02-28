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
func (s *Manager) Start(ctx context.Context, serviceID string, env []string) error {
	service, err := model.LoadService(ctx, serviceID)
	if err != nil {
		return err
	}

	for _, depID := range service.DependenciesIDs {
		if err := s.startService(ctx, depID, env); err != nil {
			return err
		}
	}

	return nil
}

// Stop stops a running Service.
func (s *Manager) Stop(ctx context.Context, serviceID string) error {
	return model.LockServiceE(ctx, serviceID, func(service *model.Service) error {
		if service.Status != model.ServiceStatusRunning {
			return model.ErrNotRunning
		}

		service.Status = model.ServiceStatusStopping
		service.MustStore(ctx)

		actual, ok := s.commands.Load(serviceID)
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

// Clean terminates all running Services.
func (s *Manager) Clean(ctx context.Context) {
	appCtx := appcontext.Get(ctx)
	lastMsgID := appCtx.Subs.LastMessageID()
	waitGroup := sync.WaitGroup{}

	s.commands.Range(func(k, _ interface{}) bool {
		serviceID := k.(string)

		appCtx.Log.DebugWithOwner(ctx, serviceID, "stopping service")

		if err := s.Stop(ctx, serviceID); err != nil {
			appCtx.Log.ErrorWithOwner(
				ctx,
				serviceID,
				"failed to stop service because %s",
				err.Error(),
			)
			return true
		}

		waitGroup.Add(1)

		serviceCtx, cancel := context.WithCancel(ctx)

		go func() {
			<-serviceCtx.Done()
			waitGroup.Done()
		}()

		appCtx.Subs.Subscribe(serviceCtx, model.MessageTypeServiceStored, lastMsgID, func(msg interface{}) {
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

		return true
	})

	waitGroup.Wait()
}

func (s *Manager) startService(ctx context.Context, serviceID string, env []string) error {
	appCtx := appcontext.Get(ctx)

	return model.LockServiceE(ctx, serviceID, func(service *model.Service) error {
		switch service.Status {
		case model.ServiceStatusStarting, model.ServiceStatusStopping:
			return model.ErrNotStopped
		case model.ServiceStatusRunning:
			return nil
		case model.ServiceStatusStopped:
			atomic.AddInt64(&s.stoppedCounter, -1)
		case model.ServiceStatusFailed:
			atomic.AddInt64(&s.failedCounter, -1)
		}

		service.Status = model.ServiceStatusStarting
		service.MustStore(ctx)
		atomic.AddInt64(&s.startingCounter, 1)
		s.updateMetrics(ctx)

		cmd := exec.Command("bash", "-l", "-c", service.Command)

		workspace := service.Workspace(ctx)
		ownerID := workspace.ID

		if project := service.Project(ctx); project != nil {
			ownerID = project.ID
			cmd.Dir = appCtx.GetProjectPath(workspace.Slug, project.Slug)
		}

		stdout := util.CreateLineWriter(ctx, appCtx.Log.InfoWithOwner, ownerID)
		stderr := util.CreateLineWriter(ctx, appCtx.Log.WarningWithOwner, ownerID)

		cmd.Env = env
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		for _, taskID := range service.BeforeIDs {
			err := model.MustLockTaskE(ctx, taskID, func(task *model.Task) error {
				return task.Run(ctx, env)
			})
			if err != nil {
				service.Status = model.ServiceStatusFailed
				service.MustStore(ctx)
				atomic.AddInt64(&s.startingCounter, -1)
				atomic.AddInt64(&s.failedCounter, 1)
				s.updateMetrics(ctx)
			}
		}

		appCtx.Log.InfoWithOwner(ctx, ownerID, service.Command)

		err := cmd.Start()
		if err == nil {
			service.Status = model.ServiceStatusRunning
			atomic.AddInt64(&s.runningCounter, 1)
		} else {
			service.Status = model.ServiceStatusFailed
			atomic.AddInt64(&s.failedCounter, 1)
		}

		service.MustStore(ctx)
		atomic.AddInt64(&s.startingCounter, -1)
		s.updateMetrics(ctx)

		if err != nil {
			appCtx.Log.ErrorWithOwner(
				ctx,
				ownerID,
				"service failed because %s",
				err.Error(),
			)
			stdout.Close()
			stderr.Close()

			return err
		}

		s.commands.Store(serviceID, cmd)

		go func() {
			err := cmd.Wait()

			model.MustLockService(ctx, serviceID, func(service *model.Service) {
				s.commands.Delete(serviceID)

				for _, taskID := range service.AfterIDs {
					taskError := model.MustLockTaskE(ctx, taskID, func(task *model.Task) error {
						return task.Run(ctx, env)
					})
					if err == nil {
						err = taskError
					}
				}

				if err == nil {
					service.Status = model.ServiceStatusStopped
					atomic.AddInt64(&s.stoppedCounter, 1)
					appCtx.Log.DebugWithOwner(ctx, ownerID, "service done")
				} else {
					service.Status = model.ServiceStatusFailed
					atomic.AddInt64(&s.failedCounter, 1)
					appCtx.Log.ErrorWithOwner(
						ctx,
						ownerID,
						"service failed because %s",
						err.Error(),
					)
				}

				atomic.AddInt64(&s.runningCounter, -1)
				service.MustStore(ctx)
			})

			s.updateMetrics(ctx)

			stdout.Close()
			stderr.Close()
		}()

		return nil
	})
}

func (s *Manager) updateMetrics(ctx context.Context) {
	appCtx := appcontext.Get(ctx)
	system := model.MustLoadSystem(ctx, appCtx.SystemID)

	model.MustLockServiceMetrics(ctx, system.ServiceMetricsID, func(metrics *model.ServiceMetrics) {
		metrics.Stopped = int(atomic.LoadInt64(&s.stoppedCounter))
		metrics.Starting = int(atomic.LoadInt64(&s.startingCounter))
		metrics.Running = int(atomic.LoadInt64(&s.runningCounter))
		metrics.Stopping = int(atomic.LoadInt64(&s.stoppingCounter))
		metrics.Failed = int(atomic.LoadInt64(&s.failedCounter))
		metrics.MustStore(ctx)
	})
}

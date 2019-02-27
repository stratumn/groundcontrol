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
	"bufio"
	"context"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ServiceManager manages creating and running job.
type ServiceManager struct {
	commands sync.Map

	stoppedCounter  int64
	startingCounter int64
	runningCounter  int64
	stoppingCounter int64
	failedCounter   int64
}

// NewServiceManager creates a ServiceManager.
func NewServiceManager() *ServiceManager {
	return &ServiceManager{}
}

// Start starts a Service and its dependencies.
func (s *ServiceManager) Start(ctx context.Context, serviceID string, env []string) error {
	service, err := LoadService(ctx, serviceID)
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
func (s *ServiceManager) Stop(ctx context.Context, serviceID string) error {
	return LockServiceE(ctx, serviceID, func(service *Service) error {
		if service.Status != ServiceStatusRunning {
			return ErrNotRunning
		}

		service.Status = ServiceStatusStopping
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
func (s *ServiceManager) Clean(ctx context.Context) {
	modelCtx := GetContext(ctx)
	lastMsgID := modelCtx.Subs.LastMessageID()
	waitGroup := sync.WaitGroup{}

	s.commands.Range(func(k, _ interface{}) bool {
		serviceID := k.(string)

		modelCtx.Log.DebugWithOwner(ctx, serviceID, "stopping service")

		if err := s.Stop(ctx, serviceID); err != nil {
			modelCtx.Log.ErrorWithOwner(
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

		modelCtx.Subs.Subscribe(serviceCtx, MessageTypeServiceStored, lastMsgID, func(msg interface{}) {
			id := msg.(string)
			if id != serviceID {
				return
			}

			service := MustLoadService(ctx, id)

			switch service.Status {
			case ServiceStatusStopped, ServiceStatusFailed:
				cancel()
			}
		})

		return true
	})

	waitGroup.Wait()
}

func (s *ServiceManager) startService(ctx context.Context, serviceID string, env []string) error {
	modelCtx := GetContext(ctx)

	return LockServiceE(ctx, serviceID, func(service *Service) error {
		switch service.Status {
		case ServiceStatusStarting, ServiceStatusStopping:
			return ErrNotStopped
		case ServiceStatusRunning:
			return nil
		case ServiceStatusStopped:
			atomic.AddInt64(&s.stoppedCounter, -1)
		case ServiceStatusFailed:
			atomic.AddInt64(&s.failedCounter, -1)
		}

		service.Status = ServiceStatusStarting
		service.MustStore(ctx)
		atomic.AddInt64(&s.startingCounter, 1)
		s.updateMetrics(ctx)

		cmd := exec.Command("bash", "-l", "-c", service.Command)

		workspace := service.Workspace(ctx)
		ownerID := workspace.ID

		if project := service.Project(ctx); project != nil {
			ownerID = project.ID
			cmd.Dir = modelCtx.GetProjectPath(workspace.Slug, project.Slug)
		}

		stdout := CreateLineWriter(ctx, modelCtx.Log.InfoWithOwner, ownerID)
		stderr := CreateLineWriter(ctx, modelCtx.Log.WarningWithOwner, ownerID)

		cmd.Env = env
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		cmd.Stdout = stdout
		cmd.Stderr = stderr

		for _, taskID := range service.BeforeIDs {
			err := MustLockTaskE(ctx, taskID, func(task *Task) error {
				return task.Run(ctx, env)
			})
			if err != nil {
				service.Status = ServiceStatusFailed
				service.MustStore(ctx)
				atomic.AddInt64(&s.startingCounter, -1)
				atomic.AddInt64(&s.failedCounter, 1)
				s.updateMetrics(ctx)
			}
		}

		modelCtx.Log.InfoWithOwner(ctx, ownerID, service.Command)

		err := cmd.Start()
		if err == nil {
			service.Status = ServiceStatusRunning
			atomic.AddInt64(&s.runningCounter, 1)
		} else {
			service.Status = ServiceStatusFailed
			atomic.AddInt64(&s.failedCounter, 1)
		}

		service.MustStore(ctx)
		atomic.AddInt64(&s.startingCounter, -1)
		s.updateMetrics(ctx)

		if err != nil {
			modelCtx.Log.ErrorWithOwner(
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

			MustLockService(ctx, serviceID, func(service *Service) {
				s.commands.Delete(serviceID)

				for _, taskID := range service.AfterIDs {
					taskError := MustLockTaskE(ctx, taskID, func(task *Task) error {
						return task.Run(ctx, env)
					})
					if err == nil {
						err = taskError
					}
				}

				if err == nil {
					service.Status = ServiceStatusStopped
					atomic.AddInt64(&s.stoppedCounter, 1)
					modelCtx.Log.DebugWithOwner(ctx, ownerID, "service done")
				} else {
					service.Status = ServiceStatusFailed
					atomic.AddInt64(&s.failedCounter, 1)
					modelCtx.Log.ErrorWithOwner(
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

func (s *ServiceManager) updateMetrics(ctx context.Context) {
	modelCtx := GetContext(ctx)
	system := MustLoadSystem(ctx, modelCtx.SystemID)

	MustLockServiceMetrics(ctx, system.ServiceMetricsID, func(metrics *ServiceMetrics) {
		metrics.Stopped = int(atomic.LoadInt64(&s.stoppedCounter))
		metrics.Starting = int(atomic.LoadInt64(&s.startingCounter))
		metrics.Running = int(atomic.LoadInt64(&s.runningCounter))
		metrics.Stopping = int(atomic.LoadInt64(&s.stoppingCounter))
		metrics.Failed = int(atomic.LoadInt64(&s.failedCounter))
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

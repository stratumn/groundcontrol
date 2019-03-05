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

package work

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"groundcontrol/appcontext"
	"groundcontrol/model"
	"groundcontrol/relay"
)

// Worker is a function that is executed once a job is running.
type Worker = func(ctx context.Context) error

// Queue queues Jobs and executes them.
type Queue struct {
	concurrency int

	hiCh    chan message
	ch      chan message
	cancels sync.Map

	lastID uint64

	mu   sync.Mutex
	done bool

	queuedCounter   int64
	runningCounter  int64
	stoppingCounter int64
	doneCounter     int64
	failedCounter   int64
}

type message struct {
	Job *model.Job
	Fn  Worker
}

// NewQueue creates a Queue with given concurrency.
// ChannelSize is the size for each channel of the queue.
// A job will fail if the channel of the corresponding priority is full.
func NewQueue(concurrency, channelSize int) *Queue {
	return &Queue{
		concurrency: concurrency,
		lastID:      uint64(time.Now().Unix()),
		ch:          make(chan message, channelSize),
		hiCh:        make(chan message, channelSize),
	}
}

// Work starts running jobs and blocks until the context is done.
func (q *Queue) Work(ctx context.Context) error {
	wg := sync.WaitGroup{}
	wg.Add(q.concurrency)
	for i := 0; i < q.concurrency; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-q.hiCh:
					q.run(ctx, msg)
					continue
				default:
				}
				select {
				case <-ctx.Done():
					return
				case msg := <-q.hiCh:
					q.run(ctx, msg)
				case msg := <-q.ch:
					q.run(ctx, msg)
				}
			}
		}()
	}
	wg.Wait()
	q.clean(ctx)
	return ctx.Err()
}

// Add adds a job to the queue and returns the job's ID.
func (q *Queue) Add(ctx context.Context, name, ownerID string, highPriority bool, fn Worker) string {
	q.mu.Lock()
	defer q.mu.Unlock()
	// Don't save job til we know if it's queued or starts running immediately.
	job := q.initJob(ctx, name, ownerID, highPriority)
	if q.done {
		q.setStatus(ctx, job, model.JobStatusFailed)
		job.MustStore(ctx)
		q.addJobToSystem(ctx, job.ID)
		appCtx := appcontext.Get(ctx)
		appCtx.Log.DebugWithOwner(ctx, appCtx.SystemID, "job failed because queue is done")
		return job.ID
	}
	msg := message{Job: job, Fn: fn}
	q.send(ctx, msg, highPriority)
	return job.ID
}

// Stop cancels a queued or running job. If it has another status state it returns ErrStatus.
func (q *Queue) Stop(ctx context.Context, id string) error {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	systemID := appCtx.SystemID
	return model.LockJobE(ctx, id, func(job *model.Job) error {
		switch job.Status {
		case model.JobStatusQueued:
			q.setStatus(ctx, job, model.JobStatusFailed)
			log.ErrorWithOwner(ctx, systemID, "job failed because it was stopped")
		case model.JobStatusRunning:
			q.setStatus(ctx, job, model.JobStatusStopping)
			log.DebugWithOwner(ctx, systemID, "job stopping")
			actual, _ := q.cancels.Load(id)
			actual.(context.CancelFunc)()
		default:
			return ErrStatus
		}
		job.UpdatedAt = model.DateTime(time.Now())
		job.MustStore(ctx)
		return nil
	})
}

func (q *Queue) send(ctx context.Context, msg message, highPriority bool) {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	systemID := appCtx.SystemID
	ch := q.ch
	if highPriority {
		ch = q.hiCh
	}
	select {
	case ch <- msg:
		model.MustLockOrNewJob(ctx, msg.Job.ID, func(_ *model.Job, isNew bool) {
			if !isNew {
				// Already stored and at least running.
				return
			}
			// Contains the job initialized in Add() but not stored yet.
			job := msg.Job
			q.setStatus(ctx, job, model.JobStatusQueued)
			job.MustStore(ctx)
			q.addJobToSystem(ctx, job.ID)
			log.DebugWithOwner(ctx, systemID, "job queued")
		})
	default:
		model.MustLockJob(ctx, msg.Job.ID, func(job *model.Job) {
			q.setStatus(ctx, job, model.JobStatusFailed)
			job.MustStore(ctx)
			q.addJobToSystem(ctx, job.ID)
			log.DebugWithOwner(ctx, systemID, "job failed because queue is full")
		})
	}
}

func (q *Queue) run(ctx context.Context, msg message) {
	var jobCtx context.Context
	var cancel context.CancelFunc
	stopped := false
	jobID := msg.Job.ID
	model.MustLockOrNewJob(ctx, jobID, func(job *model.Job, isNew bool) {
		if isNew {
			// Job isn't store yet and started running before even being queued.
			*job = *msg.Job
			q.addJobToSystem(ctx, job.ID)
		} else if job.Status == model.JobStatusFailed {
			// Job was canceled by calling Stop().
			stopped = true
			return
		}
		job.UpdatedAt = model.DateTime(time.Now())
		q.setStatus(ctx, job, model.JobStatusRunning)
		job.MustStore(ctx)
		jobCtx, cancel = q.createCtx(ctx)
		q.cancels.Store(job.ID, cancel)
	})
	if stopped {
		return
	}
	appCtx := appcontext.Get(ctx)
	appCtx.Log.DebugWithOwner(ctx, appCtx.SystemID, "job running")
	q.handleJobErr(ctx, jobID, msg.Fn(jobCtx))
	cancel()
}

func (q *Queue) handleJobErr(ctx context.Context, jobID string, err error) {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	systemID := appCtx.SystemID
	model.MustLockJob(ctx, jobID, func(job *model.Job) {
		if err != nil {
			q.setStatus(ctx, job, model.JobStatusFailed)
			log.ErrorWithOwner(ctx, systemID, "job failed because %s", err.Error())
		} else {
			q.setStatus(ctx, job, model.JobStatusDone)
			log.DebugWithOwner(ctx, systemID, "job done")
		}
		q.cancels.Delete(job.ID)
		job.UpdatedAt = model.DateTime(time.Now())
		job.MustStore(ctx)
	})
}

func (q *Queue) clean(ctx context.Context) {
	q.mu.Lock()
	q.done = true
	q.mu.Unlock()
	q.drain(ctx, q.ch)
	q.drain(ctx, q.hiCh)
}

func (q *Queue) drain(ctx context.Context, ch chan message) {
	for {
		select {
		case msg := <-ch:
			model.MustLockOrNewJob(ctx, msg.Job.ID, func(job *model.Job, isNew bool) {
				if isNew {
					*job = *msg.Job
				}
				job.UpdatedAt = model.DateTime(time.Now())
				q.setStatus(ctx, job, model.JobStatusFailed)
				job.MustStore(ctx)
			})
		default:
			close(ch)
			return
		}
	}
}

func (q *Queue) initJob(ctx context.Context, name, ownerID string, highPriority bool) *model.Job {
	priority := model.JobPriorityNormal
	if highPriority {
		priority = model.JobPriorityHigh
	}
	id := atomic.AddUint64(&q.lastID, 1)
	jobID := relay.EncodeID(model.NodeTypeJob, fmt.Sprint(id))
	now := model.DateTime(time.Now())
	return &model.Job{
		ID:        jobID,
		Priority:  priority,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		OwnerID:   ownerID,
	}
}

func (q *Queue) addJobToSystem(ctx context.Context, jobID string) {
	systemID := appcontext.Get(ctx).SystemID
	model.MustLockSystem(ctx, systemID, func(system *model.System) {
		system.JobsIDs = append([]string{jobID}, system.JobsIDs...)
		system.MustStore(ctx)
	})
}

func (q *Queue) createCtx(ctx context.Context) (context.Context, context.CancelFunc) {
	appCtx := appcontext.Get(ctx)
	ctx = appcontext.With(ctx, appCtx)
	return context.WithCancel(ctx)
}

func (q *Queue) setStatus(ctx context.Context, job *model.Job, status model.JobStatus) {
	was := job.Status
	if was == status {
		return
	}
	job.Status = status
	q.decCounter(was)
	q.incCounter(status)
	q.storeMetrics(ctx)
}

func (q *Queue) incCounter(status model.JobStatus) {
	q.addToCounter(status, 1)
}

func (q *Queue) decCounter(status model.JobStatus) {
	q.addToCounter(status, -1)
}

func (q *Queue) addToCounter(status model.JobStatus, delta int64) {
	switch status {
	case model.JobStatusQueued:
		atomic.AddInt64(&q.queuedCounter, delta)
	case model.JobStatusRunning:
		atomic.AddInt64(&q.runningCounter, delta)
	case model.JobStatusStopping:
		atomic.AddInt64(&q.stoppingCounter, delta)
	case model.JobStatusDone:
		atomic.AddInt64(&q.doneCounter, delta)
	case model.JobStatusFailed:
		atomic.AddInt64(&q.failedCounter, delta)
	}
}

func (q *Queue) storeMetrics(ctx context.Context) {
	appCtx := appcontext.Get(ctx)
	system := model.MustLoadSystem(ctx, appCtx.SystemID)
	model.MustLockJobMetrics(ctx, system.JobMetricsID, func(metrics *model.JobMetrics) {
		metrics.Queued = int(atomic.LoadInt64(&q.queuedCounter))
		metrics.Running = int(atomic.LoadInt64(&q.runningCounter))
		metrics.Stopping = int(atomic.LoadInt64(&q.stoppingCounter))
		metrics.Done = int(atomic.LoadInt64(&q.doneCounter))
		metrics.Failed = int(atomic.LoadInt64(&q.failedCounter))
		metrics.MustStore(ctx)
	})
}

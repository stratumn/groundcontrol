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
	"groundcontrol/queue"
	"groundcontrol/relay"
)

// Queue queues Jobs and executes them.
type Queue struct {
	queue   *queue.Queue
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

// NewQueue creates a Queue with given concurrency.
func NewQueue(concurrency int) *Queue {
	return &Queue{queue: queue.New(concurrency)}
}

// Work starts running jobs and blocks until the context is done.
func (q *Queue) Work(ctx context.Context) error {
	err := q.queue.Work(ctx)
	// Queue is done.
	q.mu.Lock()
	q.done = true
	q.mu.Unlock()
	q.failQueued(ctx)
	return err
}

// Add adds a job to the queue and returns the job's ID.
func (q *Queue) Add(ctx context.Context, name string, ownerID string, highPriority bool, fn func(ctx context.Context) error) string {
	q.mu.Lock()
	defer q.mu.Unlock()
	job := q.buildJob(ctx, name, ownerID, highPriority)
	log := appcontext.Get(ctx).Log
	if q.done {
		log.ErrorWithOwner(ctx, job.ID, "job failed because queue is done")
		q.setStatus(ctx, job, model.JobStatusFailed)
		job.MustStore(ctx)
		return job.ID
	}
	q.prepend(ctx, job.ID)
	log.DebugWithOwner(ctx, job.ID, "job queued")
	q.setStatus(ctx, job, model.JobStatusQueued)
	job.MustStore(ctx)
	f := func() {
		q.runJob(ctx, job.ID, fn)
	}
	if highPriority {
		go q.queue.DoHi(f)
	} else {
		go q.queue.Do(f)
	}
	return job.ID
}

// Stop cancels a running job.
func (q *Queue) Stop(ctx context.Context, id string) error {
	return model.LockJobE(ctx, id, func(job *model.Job) error {
		if job.Status != model.JobStatusRunning {
			return model.ErrNotRunning
		}
		actual, _ := q.cancels.Load(id)
		q.setStatus(ctx, job, model.JobStatusStopping)
		job.UpdatedAt = model.DateTime(time.Now())
		job.MustStore(ctx)
		actual.(context.CancelFunc)()
		return nil
	})
}

func (q *Queue) failQueued(ctx context.Context) {
	systemID := appcontext.Get(ctx).SystemID
	system := model.MustLoadSystem(ctx, systemID)
	for _, jobID := range system.JobsIDs {
		model.MustLockJob(ctx, jobID, func(job *model.Job) {
			if job.Status != model.JobStatusQueued {
				return
			}
			q.setStatus(ctx, job, model.JobStatusFailed)
			job.MustStore(ctx)
		})
	}
}

func (q *Queue) buildJob(ctx context.Context, name, ownerID string, highPriority bool) *model.Job {
	priority := model.JobPriorityNormal
	if highPriority {
		priority = model.JobPriorityHigh
	}
	id := atomic.AddUint64(&q.lastID, 1)
	relayID := relay.EncodeID(model.NodeTypeJob, fmt.Sprint(id))
	now := model.DateTime(time.Now())
	return &model.Job{
		ID:        relayID,
		Priority:  priority,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		OwnerID:   ownerID,
	}
}

func (q *Queue) prepend(ctx context.Context, jobID string) {
	systemID := appcontext.Get(ctx).SystemID
	model.MustLockSystem(ctx, systemID, func(system *model.System) {
		system.JobsIDs = append([]string{jobID}, system.JobsIDs...)
		system.MustStore(ctx)
	})
}

func (q *Queue) runJob(ctx context.Context, jobID string, fn func(context.Context) error) {
	model.MustLockJob(ctx, jobID, func(job *model.Job) {
		if job.Status != model.JobStatusQueued {
			return
		}
		appCtx := appcontext.Get(ctx)
		log := appCtx.Log
		log.DebugWithOwner(ctx, job.ID, "job running")
		// We must create a new context because the other one closes after a request.
		ctx, cancel := context.WithCancel(appcontext.With(context.Background(), appCtx))
		defer cancel()
		q.cancels.Store(job.ID, cancel)
		defer q.cancels.Delete(job.ID)
		q.setStatus(ctx, job, model.JobStatusRunning)
		job.UpdatedAt = model.DateTime(time.Now())
		job.MustStore(ctx)
		if err := fn(ctx); err != nil {
			log.ErrorWithOwner(ctx, job.ID, "job failed because %s", err.Error())
			q.setStatus(ctx, job, model.JobStatusFailed)
		} else {
			log.DebugWithOwner(ctx, job.ID, "job done")
			q.setStatus(ctx, job, model.JobStatusDone)
		}
		job.UpdatedAt = model.DateTime(time.Now())
		job.MustStore(ctx)
	})
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

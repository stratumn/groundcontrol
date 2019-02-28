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

	"groundcontrol/model"
	"groundcontrol/queue"
	"groundcontrol/relay"
)

// Queue queues Jobs and executes them.
type Queue struct {
	queue   *queue.Queue
	cancels sync.Map

	lastID uint64

	queuedCounter  int64
	runningCounter int64
	doneCounter    int64
	failedCounter  int64

	done int32
}

// NewQueue creates a Queue with given concurrency.
func NewQueue(concurrency int) *Queue {
	return &Queue{
		queue: queue.New(concurrency),
	}
}

// Work starts running jobs and blocks until the context is done.
func (q *Queue) Work(ctx context.Context) error {
	err := q.queue.Work(ctx)

	atomic.StoreInt32(&q.done, 1)

	// Fail all queued jobs.
	model.MustLockSystem(ctx, model.GetContext(ctx).SystemID, func(system *model.System) {
		for _, id := range system.JobsIDs {
			model.MustLockJob(ctx, id, func(job *model.Job) {
				if job.Status != model.JobStatusQueued {
					return
				}

				job.Status = model.JobStatusFailed
				job.MustStore(ctx)
			})
		}
	})

	return err
}

// Add adds a job to the queue and returns the job's ID.
func (q *Queue) Add(
	ctx context.Context,
	name string,
	ownerID string,
	highPriority bool,
	fn func(ctx context.Context) error,
) string {
	modelCtx := model.GetContext(ctx)
	log := modelCtx.Log

	priority := model.JobPriorityNormal
	if highPriority {
		priority = model.JobPriorityHigh
	}

	id := atomic.AddUint64(&q.lastID, 1)
	now := model.DateTime(time.Now())
	job := model.Job{
		ID:        relay.EncodeID(model.NodeTypeJob, fmt.Sprint(id)),
		Priority:  priority,
		Name:      name,
		Status:    model.JobStatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
		OwnerID:   ownerID,
	}

	if atomic.LoadInt32(&q.done) > 0 {
		job.Status = model.JobStatusFailed
	}

	job.MustStore(ctx)

	model.MustLockSystem(ctx, modelCtx.SystemID, func(system *model.System) {
		system.JobsIDs = append([]string{job.ID}, system.JobsIDs...)
		system.MustStore(ctx)
	})

	if job.Status == model.JobStatusFailed {
		log.DebugWithOwner(ctx, job.ID, "job failed because job manager is stopped")
		atomic.AddInt64(&q.failedCounter, 1)
		q.updateMetrics(ctx)

		return job.ID
	}

	log.DebugWithOwner(ctx, job.ID, "job queued")
	atomic.AddInt64(&q.queuedCounter, 1)
	q.updateMetrics(ctx)

	do := q.queue.Do
	if priority == model.JobPriorityHigh {
		do = q.queue.DoHi
	}

	go do(func() {
		log.DebugWithOwner(ctx, job.ID, "job running")

		// We must create a new context because the other one closes after a request.
		ctx := model.WithContext(context.Background(), modelCtx)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		q.cancels.Store(job.ID, cancel)
		defer q.cancels.Delete(job.ID)

		job.Status = model.JobStatusRunning
		job.UpdatedAt = model.DateTime(time.Now())
		job.MustStore(ctx)
		atomic.AddInt64(&q.runningCounter, 1)
		atomic.AddInt64(&q.queuedCounter, -1)
		q.updateMetrics(ctx)

		if err := fn(ctx); err != nil {
			log.ErrorWithOwner(ctx, job.ID, "job failed because %s", err.Error())
			job.Status = model.JobStatusFailed
			atomic.AddInt64(&q.failedCounter, 1)
		} else {
			log.DebugWithOwner(ctx, job.ID, "job done")
			job.Status = model.JobStatusDone
			atomic.AddInt64(&q.doneCounter, 1)
		}

		job.UpdatedAt = model.DateTime(time.Now())
		job.MustStore(ctx)

		atomic.AddInt64(&q.runningCounter, -1)
		q.updateMetrics(ctx)
	})

	return job.ID
}

// Stop cancels a running job.
func (q *Queue) Stop(ctx context.Context, id string) error {
	return model.LockJobE(ctx, id, func(job *model.Job) error {
		if job.Status != model.JobStatusRunning {
			return model.ErrNotRunning
		}

		actual, ok := q.cancels.Load(id)
		if !ok {
			panic("could not find cancel function for job")
		}

		job.Status = model.JobStatusStopping
		job.UpdatedAt = model.DateTime(time.Now())
		job.MustStore(ctx)

		cancel := actual.(context.CancelFunc)
		cancel()

		return nil
	})
}

func (q *Queue) updateMetrics(ctx context.Context) {
	modelCtx := model.GetContext(ctx)
	system := model.MustLoadSystem(ctx, modelCtx.SystemID)

	model.MustLockJobMetrics(ctx, system.JobMetricsID, func(metrics *model.JobMetrics) {
		metrics.Queued = int(atomic.LoadInt64(&q.queuedCounter))
		metrics.Running = int(atomic.LoadInt64(&q.runningCounter))
		metrics.Done = int(atomic.LoadInt64(&q.doneCounter))
		metrics.Failed = int(atomic.LoadInt64(&q.failedCounter))
		metrics.MustStore(ctx)
	})
}

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
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"groundcontrol/queue"
	"groundcontrol/relay"
)

// JobManager manages creating and running job.
type JobManager struct {
	queue   *queue.Queue
	cancels sync.Map

	lastID uint64

	queuedCounter  int64
	runningCounter int64
	doneCounter    int64
	failedCounter  int64

	done int32
}

// NewJobManager creates a JobManager with given concurrency.
func NewJobManager(concurrency int) *JobManager {
	return &JobManager{
		queue: queue.New(concurrency),
	}
}

// Work starts running jobs and blocks until the context is done.
func (j *JobManager) Work(ctx context.Context) error {
	err := j.queue.Work(ctx)

	atomic.StoreInt32(&j.done, 1)

	// Fail all queued jobs.
	MustLockSystem(ctx, GetContext(ctx).SystemID, func(system *System) {
		for _, id := range system.JobsIDs {
			MustLockJob(ctx, id, func(job *Job) {
				if job.Status != JobStatusQueued {
					return
				}

				job.Status = JobStatusFailed
				job.MustStore(ctx)
			})
		}
	})

	return err
}

// Add adds a job to the queue and returns the job's ID.
func (j *JobManager) Add(
	ctx context.Context,
	name string,
	ownerID string,
	priority JobPriority,
	fn func(ctx context.Context) error,
) string {
	modelCtx := GetContext(ctx)
	log := modelCtx.Log

	id := atomic.AddUint64(&j.lastID, 1)
	now := DateTime(time.Now())
	job := Job{
		ID:        relay.EncodeID(NodeTypeJob, fmt.Sprint(id)),
		Priority:  priority,
		Name:      name,
		Status:    JobStatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
		OwnerID:   ownerID,
	}

	if atomic.LoadInt32(&j.done) > 0 {
		job.Status = JobStatusFailed
	}

	job.MustStore(ctx)

	MustLockSystem(ctx, modelCtx.SystemID, func(system *System) {
		system.JobsIDs = append([]string{job.ID}, system.JobsIDs...)
		system.MustStore(ctx)
	})

	if job.Status == JobStatusFailed {
		log.DebugWithOwner(ctx, job.ID, "job failed because job manager is stopped")
		atomic.AddInt64(&j.failedCounter, 1)
		j.updateMetrics(ctx)

		return job.ID
	}

	log.DebugWithOwner(ctx, job.ID, "job queued")
	atomic.AddInt64(&j.queuedCounter, 1)
	j.updateMetrics(ctx)

	do := j.queue.Do
	if priority == JobPriorityHigh {
		do = j.queue.DoHi
	}

	go do(func() {
		log.DebugWithOwner(ctx, job.ID, "job running")

		// We must create a new context because the other one closes after a request.
		ctx := WithContext(context.Background(), modelCtx)
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		j.cancels.Store(job.ID, cancel)
		defer j.cancels.Delete(job.ID)

		job.Status = JobStatusRunning
		job.UpdatedAt = DateTime(time.Now())
		job.MustStore(ctx)
		atomic.AddInt64(&j.runningCounter, 1)
		atomic.AddInt64(&j.queuedCounter, -1)
		j.updateMetrics(ctx)

		if err := fn(ctx); err != nil {
			log.ErrorWithOwner(ctx, job.ID, "job failed because %s", err.Error())
			job.Status = JobStatusFailed
			atomic.AddInt64(&j.failedCounter, 1)
		} else {
			log.DebugWithOwner(ctx, job.ID, "job done")
			job.Status = JobStatusDone
			atomic.AddInt64(&j.doneCounter, 1)
		}

		job.UpdatedAt = DateTime(time.Now())
		job.MustStore(ctx)

		atomic.AddInt64(&j.runningCounter, -1)
		j.updateMetrics(ctx)
	})

	return job.ID
}

// Stop cancels a running job.
func (j *JobManager) Stop(ctx context.Context, id string) error {
	return LockJobE(ctx, id, func(job *Job) error {
		if job.Status != JobStatusRunning {
			return ErrNotRunning
		}

		actual, ok := j.cancels.Load(id)
		if !ok {
			panic("could not find cancel function for job")
		}

		job.Status = JobStatusStopping
		job.UpdatedAt = DateTime(time.Now())
		job.MustStore(ctx)

		cancel := actual.(context.CancelFunc)
		cancel()

		return nil
	})
}

func (j *JobManager) updateMetrics(ctx context.Context) {
	modelCtx := GetContext(ctx)
	system := MustLoadSystem(ctx, modelCtx.SystemID)

	MustLockJobMetrics(ctx, system.JobMetricsID, func(metrics *JobMetrics) {
		metrics.Queued = int(atomic.LoadInt64(&j.queuedCounter))
		metrics.Running = int(atomic.LoadInt64(&j.runningCounter))
		metrics.Done = int(atomic.LoadInt64(&j.doneCounter))
		metrics.Failed = int(atomic.LoadInt64(&j.failedCounter))
		metrics.MustStore(ctx)
	})
}

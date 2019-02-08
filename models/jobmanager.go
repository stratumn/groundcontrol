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
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/stratumn/groundcontrol/queue"
	"github.com/stratumn/groundcontrol/relay"
)

// JobManager manages creating and running jobs.
type JobManager struct {
	queue   *queue.Queue
	cancels sync.Map

	lastID uint64

	queuedCounter  int64
	runningCounter int64
	doneCounter    int64
	failedCounter  int64
}

// NewJobManager creates a JobManager with given concurrency.
func NewJobManager(concurrency int) *JobManager {
	return &JobManager{
		queue: queue.New(concurrency),
	}
}

// Work starts running jobs and blocks until the context is done.
func (j *JobManager) Work(ctx context.Context) error {
	return j.queue.Work(ctx)
}

// Add adds a job to the queue and returns the job's ID.
func (j *JobManager) Add(
	modelCtx *ModelContext,
	name string,
	ownerID string,
	priority JobPriority,
	fn func(ctx context.Context) error,
) string {
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

	modelCtx.Log.DebugWithOwner(job.ID, "job queued")
	modelCtx.Nodes.MustStoreJob(job)
	modelCtx.Subs.Publish(JobUpserted, job.ID)

	modelCtx.Nodes.MustLockSystem(modelCtx.SystemID, func(system System) {
		system.JobIDs = append([]string{job.ID}, system.JobIDs...)
		modelCtx.Nodes.MustStoreSystem(system)
	})

	atomic.AddInt64(&j.queuedCounter, 1)
	j.publishMetrics(modelCtx)

	do := j.queue.Do
	if priority == JobPriorityHigh {
		do = j.queue.DoHi
	}

	go do(func() {
		modelCtx.Log.DebugWithOwner(job.ID, "job running")

		ctx, cancel := context.WithCancel(WithModelContext(context.Background(), modelCtx))
		defer cancel()

		j.cancels.Store(job.ID, cancel)
		defer j.cancels.Delete(job.ID)

		job.Status = JobStatusRunning
		job.UpdatedAt = DateTime(time.Now())
		modelCtx.Nodes.MustStoreJob(job)
		modelCtx.Subs.Publish(JobUpserted, job.ID)
		atomic.AddInt64(&j.runningCounter, 1)
		atomic.AddInt64(&j.queuedCounter, -1)
		j.publishMetrics(modelCtx)

		if err := fn(ctx); err != nil {
			modelCtx.Log.ErrorWithOwner(job.ID, "job failed because %s", err.Error())
			job.Status = JobStatusFailed
			atomic.AddInt64(&j.failedCounter, 1)
		} else {
			modelCtx.Log.DebugWithOwner(job.ID, "job done")
			job.Status = JobStatusDone
			atomic.AddInt64(&j.doneCounter, 1)
		}

		job.UpdatedAt = DateTime(time.Now())
		modelCtx.Nodes.MustStoreJob(job)

		modelCtx.Subs.Publish(JobUpserted, job.ID)
		atomic.AddInt64(&j.runningCounter, -1)
		j.publishMetrics(modelCtx)
	})

	return job.ID
}

// Stop cancels a running job.
func (j *JobManager) Stop(modelCtx *ModelContext, id string) error {
	return modelCtx.Nodes.LockJobE(id, func(job Job) error {
		if job.Status != JobStatusRunning {
			return ErrNotRunning
		}

		actual, ok := j.cancels.Load(id)
		if !ok {
			panic("could not find cancel function for job")
		}

		job.Status = JobStatusStopping
		job.UpdatedAt = DateTime(time.Now())
		modelCtx.Nodes.MustStoreJob(job)
		modelCtx.Subs.Publish(JobUpserted, id)

		cancel := actual.(context.CancelFunc)
		cancel()

		return nil
	})
}

func (j *JobManager) publishMetrics(modelCtx *ModelContext) {
	system := modelCtx.Nodes.MustLoadSystem(modelCtx.SystemID)

	modelCtx.Nodes.MustLockJobMetrics(system.JobMetricsID, func(metrics JobMetrics) {
		metrics.Queued = int(atomic.LoadInt64(&j.queuedCounter))
		metrics.Running = int(atomic.LoadInt64(&j.runningCounter))
		metrics.Done = int(atomic.LoadInt64(&j.doneCounter))
		metrics.Failed = int(atomic.LoadInt64(&j.failedCounter))
		modelCtx.Nodes.MustStoreJobMetrics(metrics)
	})

	modelCtx.Subs.Publish(JobMetricsUpdated, system.JobMetricsID)
}

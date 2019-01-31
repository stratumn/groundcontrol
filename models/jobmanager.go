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

	"github.com/stratumn/groundcontrol/date"
	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/queue"
	"github.com/stratumn/groundcontrol/relay"
)

// JobManager manages creating and running jobs.
type JobManager struct {
	nodes *NodeManager
	log   *Logger
	subs  *pubsub.PubSub
	queue *queue.Queue

	systemID string
	nextID   uint64

	cancels sync.Map

	queuedCounter  int64
	runningCounter int64
	doneCounter    int64
	failedCounter  int64
}

// NewJobManager creates a JobManager with given concurrency.
func NewJobManager(
	nodes *NodeManager,
	log *Logger,
	subs *pubsub.PubSub,
	concurrency int,
	systemID string,
) *JobManager {
	return &JobManager{
		nodes:    nodes,
		log:      log,
		subs:     subs,
		queue:    queue.New(concurrency),
		systemID: systemID,
	}
}

// Work starts running jobs and blocks until the context is done.
func (j *JobManager) Work(ctx context.Context) error {
	return j.queue.Work(ctx)
}

// Add adds a job to the queue and returns the job's ID.
func (j *JobManager) Add(
	name string,
	ownerID string,
	priority JobPriority,
	fn func(ctx context.Context) error,
) string {

	id := atomic.AddUint64(&j.nextID, 1)
	now := date.NowFormatted()
	job := Job{
		ID:        relay.EncodeID(NodeTypeJob, fmt.Sprint(id)),
		Priority:  priority,
		Name:      name,
		Status:    JobStatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
		OwnerID:   ownerID,
	}

	meta := struct {
		Job   Job
		Error string
	}{
		job,
		"",
	}

	j.log.Info("Job Queued", meta)
	j.nodes.MustStoreJob(job)
	j.subs.Publish(JobUpserted, job.ID)

	j.nodes.MustLockSystem(j.systemID, func(system System) {
		system.JobIDs = append([]string{job.ID}, system.JobIDs...)
		j.nodes.MustStoreSystem(system)
	})

	atomic.AddInt64(&j.queuedCounter, 1)
	j.publishMetrics()

	do := j.queue.Do
	if priority == JobPriorityHigh {
		do = j.queue.DoHi
	}

	go do(func() {
		j.log.Info("Job Running", meta)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		j.cancels.Store(job.ID, cancel)
		defer j.cancels.Delete(job.ID)

		job.Status = JobStatusRunning
		job.UpdatedAt = date.NowFormatted()
		j.nodes.MustStoreJob(job)
		j.subs.Publish(JobUpserted, job.ID)
		atomic.AddInt64(&j.runningCounter, 1)
		atomic.AddInt64(&j.queuedCounter, -1)
		j.publishMetrics()

		if err := fn(ctx); err != nil {
			meta.Error = err.Error()
			j.log.Error("Job Failed", meta)
			job.Status = JobStatusFailed
			atomic.AddInt64(&j.failedCounter, 1)
		} else {
			j.log.Info("Job Done", meta)
			job.Status = JobStatusDone
			atomic.AddInt64(&j.doneCounter, 1)
		}

		job.UpdatedAt = date.NowFormatted()
		j.nodes.MustStoreJob(job)

		j.subs.Publish(JobUpserted, job.ID)
		atomic.AddInt64(&j.runningCounter, -1)
		j.publishMetrics()
	})

	return job.ID
}

// Stop cancels a running job.
func (j *JobManager) Stop(id string) error {
	var jobError error

	err := j.nodes.LockJob(id, func(job Job) {
		if job.Status != JobStatusRunning {
			jobError = ErrNotRunning
		}

		actual, ok := j.cancels.Load(id)
		if !ok {
			panic("could not find cancel function for job")
		}

		job.Status = JobStatusStopping
		job.UpdatedAt = date.NowFormatted()
		j.nodes.MustStoreJob(job)
		j.subs.Publish(JobUpserted, id)

		cancel := actual.(context.CancelFunc)
		cancel()
	})
	if err != nil {
		return err
	}

	return jobError
}

func (j *JobManager) publishMetrics() {
	system := j.nodes.MustLoadSystem(j.systemID)

	j.nodes.MustLockJobMetrics(system.JobMetricsID, func(metrics JobMetrics) {
		metrics.Queued = int(atomic.LoadInt64(&j.queuedCounter))
		metrics.Running = int(atomic.LoadInt64(&j.runningCounter))
		metrics.Done = int(atomic.LoadInt64(&j.doneCounter))
		metrics.Failed = int(atomic.LoadInt64(&j.failedCounter))
		j.nodes.MustStoreJobMetrics(metrics)
	})

	j.subs.Publish(JobMetricsUpdated, system.JobMetricsID)
}

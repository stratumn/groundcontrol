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

	nextID uint64

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
	projectID string,
	fn func() error,
) string {
	meta := struct {
		Name      string
		ProjectID string
		Error     string
	}{
		name,
		projectID,
		"",
	}

	j.log.Info("Job Queued", meta)

	id := atomic.AddUint64(&j.nextID, 1)
	now := date.NowFormatted()
	job := Job{
		ID:        relay.EncodeID(NodeTypeJob, fmt.Sprint(id)),
		Name:      name,
		Status:    JobStatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
		ProjectID: projectID,
	}
	j.nodes.MustStoreJob(job)
	j.subs.Publish(JobUpserted, job.ID)

	j.nodes.MustLockSystem(j.systemID, func(system System) {
		system.JobIDs = append([]string{job.ID}, system.JobIDs...)
		j.nodes.MustStoreSystem(system)
	})

	atomic.AddInt64(&j.queuedCounter, 1)
	j.publishMetrics()

	go j.queue.Do(func() {
		j.log.Info("Job Running", meta)

		job.Status = JobStatusRunning
		job.UpdatedAt = date.NowFormatted()
		j.nodes.MustStoreJob(job)
		j.subs.Publish(JobUpserted, job.ID)
		atomic.AddInt64(&j.runningCounter, 1)
		atomic.AddInt64(&j.queuedCounter, -1)
		j.publishMetrics()

		if err := fn(); err != nil {
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

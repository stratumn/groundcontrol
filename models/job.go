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
	"container/list"
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/stratumn/groundcontrol/date"
	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/queue"
	"github.com/stratumn/groundcontrol/relay"
)

// Message types.
const (
	JobUpserted = "JOB_UPSERTED" // Go type *Job
)

var jobPaginator = relay.Paginator{
	GetID: func(node interface{}) string {
		return node.(*Job).ID
	},
}

// Job represents a job in the app.
type Job struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
	Status    JobStatus `json:"status"`
	Project   *Project  `json:"project"`
}

// IsNode is used by gqlgen.
func (Job) IsNode() {}

// JobManager manages creating and running jobs.
type JobManager struct {
	nodes  *NodeManager
	pubsub *pubsub.PubSub
	queue  *queue.Queue

	mu        sync.Mutex
	list      *list.List
	nextJobID uint64
}

// NewJobManager creates a JobManager with given concurrency.
func NewJobManager(nodes *NodeManager, pubsub *pubsub.PubSub, concurrency int) *JobManager {
	return &JobManager{
		nodes:  nodes,
		pubsub: pubsub,
		queue:  queue.New(concurrency),
		list:   list.New(),
	}
}

// Work starts running jobs and blocks until the context is done.
func (j *JobManager) Work(ctx context.Context) error {
	return j.queue.Work(ctx)
}

// Add adds a job to the queue.
func (j *JobManager) Add(
	name string,
	project *Project,
	fn func() error,
) *Job {
	j.mu.Lock()
	defer j.mu.Unlock()

	id := j.nextJobID
	now := date.NowFormatted()
	job := Job{
		ID:        relay.EncodeID(JobType, fmt.Sprint(id)),
		Name:      name,
		Status:    JobStatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
		Project:   project,
	}

	j.nextJobID++
	j.nodes.Store(job.ID, &job)
	j.list.PushFront(&job)
	j.pubsub.Publish(JobUpserted, &job)

	go j.queue.Do(func() {
		job.Status = JobStatusRunning
		job.UpdatedAt = date.NowFormatted()
		j.pubsub.Publish(JobUpserted, &job)

		if err := fn(); err != nil {
			log.Println(err)
			job.Status = JobStatusFailed
		} else {
			job.Status = JobStatusDone
		}

		job.UpdatedAt = date.NowFormatted()
		j.pubsub.Publish(JobUpserted, &job)
	})

	return &job
}

// Jobs returns paginated jobs and supports filtering by status.
func (j *JobManager) Jobs(
	after *string,
	before *string,
	first *int,
	last *int,
	status []JobStatus,
) (JobConnection, error) {
	jobList := list.New()
	element := j.list.Front()

	for element != nil {
		job := element.Value.(*Job)
		match := len(status) == 0

		for _, v := range status {
			if job.Status == v {
				match = true
				break
			}
		}

		if match {
			jobList.PushBack(job)
		}

		element = element.Next()
	}

	connection, err := jobPaginator.Paginate(jobList, after, before, first, last)
	if err != nil {
		return JobConnection{}, err
	}

	edges := make([]JobEdge, len(connection.Edges))

	for i, v := range connection.Edges {
		edges[i] = JobEdge{
			Node:   *v.Node.(*Job),
			Cursor: v.Cursor,
		}
	}

	return JobConnection{
		Edges:    edges,
		PageInfo: connection.PageInfo,
	}, nil
}

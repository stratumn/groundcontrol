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

package groundcontrol

import (
	"container/list"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
)

var (
	jobMu     = sync.Mutex{}
	jobList   = list.New()
	nextJobID = uint64(0)
)

var jobPaginator = Paginator{
	GetID: func(node interface{}) string {
		return node.(*Job).ID
	},
}

type Job struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
	Status    JobStatus `json:"status"`
	Project   *Project  `json:"project"`
}

func (Job) IsNode() {}

func CreateJob(name string, project *Project, fn func() error) {
	jobMu.Lock()
	defer jobMu.Unlock()

	now := NowFormatted()

	job := Job{
		ID:        EncodeID("Job", fmt.Sprint(nextJobID)),
		Name:      name,
		Status:    JobStatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
		Project:   project,
	}

	jobList.PushFront(&job)
	PublishJobUpserted(&job)
	nextJobID++

	go GlobalQueue.Do(func() {
		job.Status = JobStatusRunning
		job.UpdatedAt = NowFormatted()
		PublishJobUpserted(&job)

		if err := fn(); err != nil {
			log.Println(err)
			job.Status = JobStatusFailed
		} else {
			job.Status = JobStatusDone
		}

		job.UpdatedAt = NowFormatted()
		PublishJobUpserted(&job)
	})
}

func GetJobList() *list.List {
	return jobList
}

var (
	nextJobSubscriptionID    = uint64(0)
	jobUpsertedSubscriptions = sync.Map{}
)

func SubscribeJobUpserted(fn func(*Job)) func() {
	id := atomic.AddUint64(&nextJobSubscriptionID, 1)
	jobUpsertedSubscriptions.Store(id, fn)

	return func() {
		jobUpsertedSubscriptions.Delete(id)
	}
}

func PublishJobUpserted(job *Job) {
	jobUpsertedSubscriptions.Range(func(_, v interface{}) bool {
		fn := v.(func(*Job))
		fn(job)
		return true
	})
}

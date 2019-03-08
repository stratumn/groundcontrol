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

package job

import (
	"context"
	"sync"
	"time"

	"groundcontrol/appcontext"
	"groundcontrol/model"
)

// StartPeriodic queues Jobs periodically.
// The waitTime argument defines how long to wait after Jobs are finished before starting another round.
// The chain arguments should be a series function that creates Jobs and returns their IDs.
// The first round of Jobs will be created immediately upon calling this function.
// This function blocks until the context is canceled.
func StartPeriodic(ctx context.Context, waitTime time.Duration, chain ...func(context.Context) []string) error {
	for _, fn := range chain {
		periodicRound(ctx, fn)
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			for _, fn := range chain {
				periodicRound(ctx, fn)
			}
		}
	}
}

func periodicRound(ctx context.Context, fn func(context.Context) []string) {
	appCtx := appcontext.Get(ctx)
	lastMsgID := appCtx.Subs.LastMessageID()
	jobIDs := fn(ctx)
	if len(jobIDs) < 1 {
		return
	}
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(jobIDs))
	jobMap := sync.Map{}
	for _, jobID := range jobIDs {
		jobMap.Store(jobID, struct{}{})
	}
	subsCtx, cancel := context.WithCancel(appcontext.With(context.Background(), appCtx))
	defer cancel()
	appCtx.Subs.Subscribe(subsCtx, model.MessageTypeJobStored, lastMsgID, func(msg interface{}) {
		job := msg.(*model.Job)
		_, ok := jobMap.Load(job.ID)
		if !ok {
			return
		}
		switch job.Status {
		case model.JobStatusDone, model.JobStatusFailed:
			waitGroup.Done()
		}
	})
	waitGroup.Wait()
}

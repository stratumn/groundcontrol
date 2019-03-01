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

// StartPeriodic is used to run jobs periodically.
// The waitTime argument defines how long to wait after jobs are finished to create new ones.
// The addJobs argument should be a function that creates jobs and returns their IDs.
// The first round of jobs will be created immediately upon calling this function.
// This function blocks until the context is canceled.
func StartPeriodic(
	ctx context.Context,
	waitTime time.Duration,
	chain ...func(context.Context) []string,
) error {
	appCtx := appcontext.Get(ctx)

	round := func(fn func(context.Context) []string) {
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

	for _, fn := range chain {
		round(fn)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			for _, fn := range chain {
				round(fn)
			}
		}
	}
}

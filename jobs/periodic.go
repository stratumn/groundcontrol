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

package jobs

import (
	"context"
	"sync"
	"time"

	"groundcontrol/models"
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
	modelCtx := models.GetModelContext(ctx)

	round := func(fn func(context.Context) []string) {
		jobIDs := fn(ctx)
		if len(jobIDs) < 1 {
			return
		}

		waitGroup := sync.WaitGroup{}
		waitGroup.Add(len(jobIDs))

		jobMap := sync.Map{}

		for _, jobID := range jobIDs {
			jobMap.Store(jobID, true)
		}

		subsCtx, cancel := context.WithCancel(ctx)

		modelCtx.Subs.Subscribe(subsCtx, models.JobUpserted, 0, func(msg interface{}) {
			id := msg.(string)
			_, ok := jobMap.Load(id)
			if !ok {
				return
			}

			switch modelCtx.Nodes.MustLoadJob(id).Status {
			case models.JobStatusDone, models.JobStatusFailed:
				waitGroup.Done()
			}
		})

		waitGroup.Wait()
		cancel()
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

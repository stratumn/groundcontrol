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

package resolvers

import (
	"context"

	"github.com/stratumn/groundcontrol/jobs"
	"github.com/stratumn/groundcontrol/models"
)

func (r *mutationResolver) LoadProjectCommits(ctx context.Context, id string) (models.Job, error) {
	nodes := models.GetModelContext(ctx).Nodes

	jobID, err := jobs.LoadCommits(ctx, id, models.JobPriorityHigh)
	if err != nil {
		return models.Job{}, err
	}

	return nodes.MustLoadJob(jobID), nil
}

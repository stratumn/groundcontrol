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
	"fmt"
	"os"

	"groundcontrol/jobs"
	"groundcontrol/models"
)

func (r *mutationResolver) Run(
	ctx context.Context,
	id string,
	variables []models.VariableInput,
) (models.Job, error) {
	nodes := models.GetModelContext(ctx).Nodes
	subs := models.GetModelContext(ctx).Subs
	keys := models.GetModelContext(ctx).Keys
	viewerID := models.GetModelContext(ctx).ViewerID

	env := os.Environ()
	save := false

	for _, variable := range variables {
		env = append(env, fmt.Sprintf("%s=%s", variable.Name, variable.Value))

		if !variable.Save {
			continue
		}

		save = true

		keys.UpsertKey(nodes, subs, viewerID, models.KeyInput{
			Name:  variable.Name,
			Value: variable.Value,
		})
	}

	if save {
		if err := keys.Save(); err != nil {
			return models.Job{}, nil
		}
	}

	jobID, err := jobs.Run(ctx, id, env, models.JobPriorityHigh)
	if err != nil {
		return models.Job{}, err
	}

	return nodes.MustLoadJob(jobID), nil
}

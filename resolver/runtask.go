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

package resolver

import (
	"context"
	"fmt"
	"os"

	"groundcontrol/appcontext"
	"groundcontrol/job"
	"groundcontrol/model"
)

func (r *mutationResolver) RunTask(ctx context.Context, id string, variables []model.VariableInput) (*model.Job, error) {
	keys := appcontext.Get(ctx).Keys
	env := os.Environ()
	save := false
	for _, variable := range variables {
		env = append(env, fmt.Sprintf("%s=%s", variable.Name, variable.Value))
		if !variable.Save {
			continue
		}
		save = true
		keys.Set(ctx, variable.Name, variable.Value)
	}
	if save {
		if err := keys.Save(); err != nil {
			return nil, err
		}
	}
	jobID, err := job.RunTask(ctx, id, env, true)
	if err != nil {
		return nil, err
	}
	return model.LoadJob(ctx, jobID)
}

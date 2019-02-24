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

	"groundcontrol/jobs"
	"groundcontrol/models"
)

func (r *mutationResolver) AddDirectorySource(
	ctx context.Context,
	input models.DirectorySourceInput,
) (*models.DirectorySource, error) {
	modelCtx := models.GetModelContext(ctx)

	id := modelCtx.Sources.UpsertDirectorySource(ctx, input)

	if err := modelCtx.Sources.Save(); err != nil {
		return nil, err
	}

	_, err := jobs.LoadDirectorySource(ctx, id, models.JobPriorityHigh)
	if err != nil {
		return nil, err
	}

	return models.LoadDirectorySource(ctx, id)
}

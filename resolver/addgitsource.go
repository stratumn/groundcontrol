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

	"groundcontrol/appcontext"
	"groundcontrol/job"
	"groundcontrol/model"
)

func (r *mutationResolver) AddGitSource(ctx context.Context, input model.GitSourceInput) (*model.GitSource, error) {
	appCtx := appcontext.Get(ctx)
	id := appCtx.Sources.SetGitSource(ctx, input.Repository, input.Reference)
	if err := appCtx.Sources.Save(); err != nil {
		return nil, err
	}
	_, err := job.LoadGitSource(ctx, id, true)
	if err != nil {
		return nil, err
	}
	return model.LoadGitSource(ctx, id)
}

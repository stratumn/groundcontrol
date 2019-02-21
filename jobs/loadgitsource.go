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

	"groundcontrol/models"
)

// LoadGitSource loads the workspaces of the source and updates it.
func LoadGitSource(ctx context.Context, sourceID string, priority models.JobPriority) (string, error) {
	if err := startLoadingGitSource(ctx, sourceID); err != nil {
		return "", err
	}

	modelCtx := models.GetModelContext(ctx)

	return modelCtx.Jobs.Add(
		ctx,
		LoadGitSourceJob,
		sourceID,
		priority,
		func(ctx context.Context) error {
			return doLoadGitSource(ctx, sourceID)
		},
	), nil
}

func startLoadingGitSource(ctx context.Context, sourceID string) error {
	return models.LockGitSourceE(ctx, sourceID, func(source models.GitSource) error {
		if source.IsLoading {
			return ErrDuplicate
		}

		source.IsLoading = true
		source.MustStore(ctx)

		return nil
	})
}

func doLoadGitSource(ctx context.Context, sourceID string) error {
	return models.MustLockGitSourceE(ctx, sourceID, func(source models.GitSource) error {
		return source.Update(ctx)
	})
}

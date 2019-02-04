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

	"github.com/stratumn/groundcontrol/models"
)

// LoadGitSource loads the workspaces of the source and updates it.
func LoadGitSource(ctx context.Context, sourceID string, priority models.JobPriority) (string, error) {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	subs := modelCtx.Subs

	err := nodes.LockGitSourceE(sourceID, func(source models.GitSource) error {
		if source.IsLoading {
			return ErrDuplicate
		}

		source.IsLoading = true
		nodes.MustStoreGitSource(source)

		return nil
	})
	if err != nil {
		return "", err
	}

	subs.Publish(models.SourceUpserted, sourceID)

	jobID := modelCtx.Jobs.Add(
		models.GetModelContext(ctx),
		LoadGitSourceJob,
		sourceID,
		priority,
		func(ctx context.Context) error {
			return doLoadGitSource(ctx, sourceID)
		},
	)

	return jobID, nil
}

func doLoadGitSource(ctx context.Context, sourceID string) error {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	subs := modelCtx.Subs

	defer func() {
		nodes.MustLockGitSource(sourceID, func(source models.GitSource) {
			source.IsLoading = false
			nodes.MustStoreGitSource(source)
		})

		subs.Publish(models.SourceUpserted, sourceID)
	}()

	return nil
}

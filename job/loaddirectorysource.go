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

	"groundcontrol/model"
)

// LoadDirectorySource loads the workspaces of the source and updates it.
func LoadDirectorySource(ctx context.Context, sourceID string, priority model.JobPriority) (string, error) {
	if err := startLoadingDirectorySource(ctx, sourceID); err != nil {
		return "", err
	}

	modelCtx := model.GetModelContext(ctx)

	return modelCtx.Jobs.Add(
		ctx,
		JobNameLoadDirectorySource,
		sourceID,
		priority,
		func(ctx context.Context) error {
			return doLoadDirectorySource(ctx, sourceID)
		},
	), nil
}

func startLoadingDirectorySource(ctx context.Context, sourceID string) error {
	return model.LockDirectorySourceE(ctx, sourceID, func(source *model.DirectorySource) error {
		if source.IsLoading {
			return ErrDuplicate
		}

		source.IsLoading = true
		source.MustStore(ctx)

		return nil
	})
}

func doLoadDirectorySource(ctx context.Context, sourceID string) error {
	return model.MustLockDirectorySourceE(ctx, sourceID, func(source *model.DirectorySource) error {
		return source.Update(ctx)
	})
}
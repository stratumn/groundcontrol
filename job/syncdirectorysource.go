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

	"groundcontrol/appcontext"
	"groundcontrol/model"
)

// SyncDirectorySource queues a job to sync the Workspaces of the DirectorySource.
func SyncDirectorySource(ctx context.Context, sourceID string, highPriority bool) (string, error) {
	if err := startSyncingDirectorySource(ctx, sourceID); err != nil {
		return "", err
	}
	appCtx := appcontext.Get(ctx)
	return appCtx.Jobs.Add(ctx, JobNameSyncDirectorySource, sourceID, highPriority, func(ctx context.Context) error {
		return doSyncDirectorySource(ctx, sourceID)
	}), nil
}

func startSyncingDirectorySource(ctx context.Context, sourceID string) error {
	return model.LockDirectorySourceE(ctx, sourceID, func(source *model.DirectorySource) error {
		if source.IsSyncing {
			return ErrDuplicate
		}
		source.IsSyncing = true
		source.MustStore(ctx)
		return nil
	})
}

func doSyncDirectorySource(ctx context.Context, sourceID string) error {
	return model.MustLockDirectorySourceE(ctx, sourceID, func(source *model.DirectorySource) error {
		return source.Sync(ctx)
	})
}

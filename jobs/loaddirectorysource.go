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
	"os"
	"path/filepath"

	"groundcontrol/models"
)

// LoadDirectorySource loads the workspaces of the source and updates it.
func LoadDirectorySource(ctx context.Context, sourceID string, priority models.JobPriority) (string, error) {
	modelCtx := models.GetModelContext(ctx)
	subs := modelCtx.Subs

	err := models.LockDirectorySourceE(ctx, sourceID, func(source models.DirectorySource) error {
		if source.IsLoading {
			return ErrDuplicate
		}

		source.IsLoading = true
		source.MustStore(ctx)

		return nil
	})
	if err != nil {
		return "", err
	}

	subs.Publish(models.SourceUpserted, sourceID)

	jobID := modelCtx.Jobs.Add(
		ctx,
		LoadDirectorySourceJob,
		sourceID,
		priority,
		func(ctx context.Context) error {
			return doLoadDirectorySource(ctx, sourceID)
		},
	)

	return jobID, nil
}

func doLoadDirectorySource(ctx context.Context, sourceID string) error {
	var (
		workspaceIDs []string
		err          error
	)

	modelCtx := models.GetModelContext(ctx)
	subs := modelCtx.Subs

	defer func() {
		models.MustLockDirectorySource(ctx, sourceID, func(source models.DirectorySource) {
			if err == nil {
				source.WorkspaceIDs = workspaceIDs
			}

			source.IsLoading = false
			source.MustStore(ctx)
		})

		subs.Publish(models.SourceUpserted, sourceID)
	}()

	source := models.MustLoadDirectorySource(ctx, sourceID)

	workspaceIDs, err = walkSourceDirectory(ctx, source.Directory)

	return err
}

func walkSourceDirectory(ctx context.Context, directory string) (workspaceIDs []string, err error) {
	err = filepath.Walk(
		directory,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if info.IsDir() && info.Name() == ".git" {
				return filepath.SkipDir
			}

			if filepath.Ext(path) != ".yml" {
				return nil
			}

			config, err := models.LoadWorkspacesConfigYAML(path)
			if err != nil {
				return err
			}

			ids, err := config.UpsertNodes(ctx)
			if err != nil {
				return err
			}

			workspaceIDs = append(workspaceIDs, ids...)

			return nil
		},
	)

	return
}

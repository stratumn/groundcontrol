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

package models

import (
	"context"
	"os"
	"path/filepath"

	"groundcontrol/relay"
)

// Source is a collection of workspaces.
type Source interface {
	Node

	IsSource()

	// User returns the user who owns the source.
	User(context.Context) User

	// GetWorkspacesIDs returns the IDs of the workspaces.
	GetWorkspacesIDs() []string

	// Workspaces are the workspaces using Relay pagination.
	Workspaces(
		ctx context.Context,
		after *string,
		before *string,
		first *int,
		last *int,
	) (WorkspaceConnection, error)
}

// MustLoadSource loads a Source or panics on failure.
func MustLoadSource(ctx context.Context, id string) Source {
	identifiers, err := relay.DecodeID(id)
	if err != nil {
		panic(err)
	}

	switch identifiers[0] {
	case NodeTypeDirectorySource:
		return MustLoadDirectorySource(ctx, id)
	case NodeTypeGitSource:
		return MustLoadGitSource(ctx, id)
	}

	panic(ErrType)
}

// LoadWorkspacesInSource loads and stores nodes for the workspaces in a directory recursively.
func LoadWorkspacesInSource(
	ctx context.Context,
	directory string,
	sourceID string,
) ([]string, error) {
	var workspaceIDs []string

	return workspaceIDs, filepath.Walk(
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

			config, err := LoadWorkspacesConfigYAML(path)
			if err != nil {
				return err
			}

			ids, err := config.UpsertNodes(ctx, sourceID)
			if err != nil {
				return err
			}

			workspaceIDs = append(workspaceIDs, ids...)

			return nil
		},
	)
}

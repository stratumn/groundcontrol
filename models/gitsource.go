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

	"groundcontrol/util"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// GetWorkspacesIDs returns the IDs of the workspaces.
func (n GitSource) GetWorkspacesIDs() []string {
	return n.WorkspacesIDs
}

// IsCloned checks if the project is cloned.
func (n GitSource) IsCloned(ctx context.Context) bool {
	getGitSourcePath := GetModelContext(ctx).GetGitSourcePath
	directory := getGitSourcePath(n.Repository, n.Reference)

	return util.FileExists(directory)
}

// ReferenceShort returns the short name of the reference.
func (n GitSource) ReferenceShort() string {
	return plumbing.ReferenceName(n.Reference).Short()
}

// Path returns the path to the source.
func (n GitSource) Path(ctx context.Context) string {
	modelCtx := GetModelContext(ctx)
	return modelCtx.GetGitSourcePath(n.Repository, n.Reference)
}

// Update loads the latest commits and upserts the source.
func (n *GitSource) Update(ctx context.Context) error {
	n.IsLoading = true
	n.MustStore(ctx)

	defer func() {
		n.IsLoading = false
		n.MustStore(ctx)
	}()

	if err := n.pullOrClone(ctx); err != nil {
		return err
	}

	workspaceIDs, err := LoadWorkspacesInSource(ctx, n.Path(ctx), n.ID)
	if err != nil {
		return err
	}

	n.WorkspacesIDs = workspaceIDs
	n.MustStore(ctx)

	return nil
}

// pullOrClone pulls the directory if already cloned, otherwise it clones it.
func (n GitSource) pullOrClone(ctx context.Context) error {
	if n.IsCloned(ctx) {
		return n.pull(ctx)
	}

	return n.clone(ctx)
}

// clone clones the remote repository.
func (n GitSource) clone(ctx context.Context) error {
	options := &git.CloneOptions{
		URL:           n.Repository,
		ReferenceName: plumbing.ReferenceName(n.Reference),
	}

	_, err := git.PlainCloneContext(ctx, n.Path(ctx), false, options)

	return err
}

// pull pulls the remote repository.
func (n GitSource) pull(ctx context.Context) error {
	repo, err := n.openRepository(ctx)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	options := &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.ReferenceName(n.Reference),
	}

	err = worktree.PullContext(ctx, options)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}

	return err
}

// openRepository opens the repository of the source.
func (n GitSource) openRepository(ctx context.Context) (*git.Repository, error) {
	return git.PlainOpen(n.Path(ctx))
}

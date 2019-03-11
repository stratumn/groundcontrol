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

package model

import (
	"context"
	"path/filepath"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"

	"groundcontrol/appcontext"
	"groundcontrol/util"
)

// String is a string representation for the type instance.
func (n *GitSource) String() string {
	return filepath.Base(n.Repository)
}

// ReferenceShort is the short name of the Reference.
func (n *GitSource) ReferenceShort() string {
	return plumbing.ReferenceName(n.Reference).Short()
}

// IsCloned indicates whether the repository is cloned.
func (n *GitSource) IsCloned(ctx context.Context) bool {
	getGitSourcePath := appcontext.Get(ctx).GetGitSourcePath
	directory := getGitSourcePath(n.Repository, n.Reference)
	return util.FileExists(directory)
}

// Path returns the path to the source.
func (n *GitSource) Path(ctx context.Context) string {
	appCtx := appcontext.Get(ctx)
	return appCtx.GetGitSourcePath(n.Repository, n.Reference)
}

// Sync syncs the Source.
func (n *GitSource) Sync(ctx context.Context) error {
	defer func() {
		n.IsSyncing = false
		n.MustStore(ctx)
	}()
	n.IsSyncing = true
	n.MustStore(ctx)

	if err := n.pullOrClone(ctx); err != nil {
		return err
	}
	workspaceIDs, err := SyncWorkspacesInDirectory(ctx, n.Path(ctx), n.ID)
	if err != nil {
		return err
	}
	n.WorkspacesIDs = workspaceIDs
	n.MustStore(ctx)
	return nil
}

// pullOrClone pulls the directory if already cloned, otherwise it clones it.
func (n *GitSource) pullOrClone(ctx context.Context) error {
	if n.IsCloned(ctx) {
		return n.pull(ctx)
	}
	return n.clone(ctx)
}

// clone clones the remote repository.
func (n *GitSource) clone(ctx context.Context) error {
	opts := git.CloneOptions{
		URL:           n.Repository,
		ReferenceName: plumbing.ReferenceName(n.Reference),
	}
	_, err := git.PlainCloneContext(ctx, n.Path(ctx), false, &opts)
	return err
}

// pull pulls the remote repository.
func (n *GitSource) pull(ctx context.Context) error {
	repo, err := n.openRepository(ctx)
	if err != nil {
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	opts := git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.ReferenceName(n.Reference),
	}
	err = worktree.PullContext(ctx, &opts)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	return err
}

// openRepository opens the repository of the source.
func (n *GitSource) openRepository(ctx context.Context) (*git.Repository, error) {
	return git.PlainOpen(n.Path(ctx))
}

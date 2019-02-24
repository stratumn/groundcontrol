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
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"groundcontrol/gitutil"
	"groundcontrol/relay"
	"groundcontrol/util"
)

// ReferenceShort is he short name of the Reference.
func (n *Project) ReferenceShort() string {
	return plumbing.ReferenceName(n.Reference).Short()
}

// RemoteReferenceShort is he short name of the RemoteReference.
func (n *Project) RemoteReferenceShort() string {
	return plumbing.ReferenceName(n.RemoteReference).Short()
}

// LocalReferenceShort is he short name of the LocalReference.
func (n *Project) LocalReferenceShort() string {
	return plumbing.ReferenceName(n.LocalReference).Short()
}

// IsCloned indicates whether the repository is cloned.
func (n *Project) IsCloned(ctx context.Context) bool {
	return util.FileExists(n.Path(ctx))
}

// IsCached checks if the project is cached.
func (n *Project) IsCached(ctx context.Context) bool {
	return util.FileExists(n.CachePath(ctx))
}

// Path returns the path to the project.
func (n *Project) Path(ctx context.Context) string {
	modelCtx := GetModelContext(ctx)
	return modelCtx.GetProjectPath(n.Workspace(ctx).Slug, n.Slug)
}

// CachePath returns the path to the project's cache.
func (n *Project) CachePath(ctx context.Context) string {
	modelCtx := GetModelContext(ctx)
	return modelCtx.GetProjectCachePath(n.Workspace(ctx).Slug, n.Slug)
}

// Clone clones and upserts the project.
func (n *Project) Clone(ctx context.Context) error {
	n.IsCloning = true
	n.MustStore(ctx)

	defer func() {
		n.IsCloning = false
		n.MustStore(ctx)
	}()

	path := n.Path(ctx)
	options := &git.CloneOptions{
		URL:           n.Repository,
		ReferenceName: plumbing.ReferenceName(n.RemoteReference),
	}

	if _, err := git.PlainCloneContext(ctx, path, false, options); err != nil {
		return err
	}

	return n.Update(ctx)
}

// Pull pulls and upserts the project.
func (n *Project) Pull(ctx context.Context) error {
	n.IsPulling = true
	n.MustStore(ctx)

	defer func() {
		n.IsPulling = false
		n.MustStore(ctx)
	}()

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
		ReferenceName: plumbing.ReferenceName(n.RemoteReference),
	}

	err = worktree.PullContext(ctx, options)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	if err != nil {
		return err
	}

	return n.Update(ctx)
}

// Update loads the latest commits and upserts the project.
func (n *Project) Update(ctx context.Context) error {
	n.IsLoadingCommits = true
	n.MustStore(ctx)

	defer func() {
		n.IsLoadingCommits = false
		n.MustStore(ctx)
	}()

	if !n.IsCloned(ctx) && !n.IsCached(ctx) {
		if err := n.cloneCache(ctx); err != nil {
			return err
		}
	} else {
		if err := n.fetch(ctx); err != nil {
			return nil
		}
	}

	if err := n.updateReferences(ctx); err != nil {
		return err
	}

	remoteCommitsIDs, err := n.loadCommits(ctx, n.localRemoteReferenceName())
	if err != nil {
		return err
	}

	localCommitsIDs, err := n.loadCommits(ctx, plumbing.ReferenceName(n.LocalReference))
	if err != nil {
		return err
	}

	if len(remoteCommitsIDs) > 0 {
		n.RemoteCommitsIDs = remoteCommitsIDs
	}

	if len(localCommitsIDs) > 0 {
		n.LocalCommitsIDs = localCommitsIDs
	}

	return n.updateStatus(ctx)
}

// openRepository opens the repository of the project.
// If the project isn't cloned it will be a bare repository.
// It returns nil if the project isn't cloned or cached.
func (n *Project) openRepository(ctx context.Context) (*git.Repository, error) {
	if n.IsCloned(ctx) {
		return git.PlainOpen(n.Path(ctx))
	}

	if n.IsCached(ctx) {
		return git.PlainOpen(n.CachePath(ctx))
	}

	return nil, nil
}

// cloneCache bare clones the repository into the cache.
func (n *Project) cloneCache(ctx context.Context) error {
	cachePath := n.CachePath(ctx)
	options := &git.CloneOptions{
		URL:           n.Repository,
		ReferenceName: plumbing.ReferenceName(n.RemoteReference),
	}

	_, err := git.PlainCloneContext(ctx, cachePath, true, options)
	return err
}

// fetch fetches either the cloned or the cached repository.
func (n *Project) fetch(ctx context.Context) error {
	repo, err := n.openRepository(ctx)
	if err != nil {
		return err
	}

	options := &git.FetchOptions{
		Force: !n.IsCloned(ctx),
	}

	err = repo.FetchContext(ctx, options)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	return nil
}

// loadCommits loads the commits of a reference.
func (n *Project) loadCommits(
	ctx context.Context,
	refName plumbing.ReferenceName,
) ([]string, error) {
	var commitIDs []string

	iter, err := n.iterateCommits(ctx, refName)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	return commitIDs, iter.ForEach(func(c *object.Commit) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		commit := Commit{
			ID:       relay.EncodeID(NodeTypeCommit, c.Hash.String()),
			Hash:     Hash(c.Hash.String()),
			Headline: strings.Split(c.Message, "\n")[0],
			Message:  c.Message,
			Author:   c.Author.Name,
			Date:     DateTime(c.Author.When),
		}

		commit.MustStore(ctx)
		commitIDs = append(commitIDs, commit.ID)

		return nil
	})
}

// updateReferences sets the local and remote references according to the current branch.
func (n *Project) updateReferences(ctx context.Context) error {
	branch, err := n.currentBranch(ctx)
	if err != nil {
		return err
	}

	if branch != nil {
		n.RemoteReference = branch.Merge.String()
		n.LocalReference = plumbing.NewBranchReferenceName(branch.Name).String()

	} else {
		n.RemoteReference = n.Reference
		n.LocalReference = n.Reference
	}

	return nil
}

// updateStatus updates the status of the project according to the local branch.
func (n *Project) updateStatus(ctx context.Context) error {
	repo, err := n.openRepository(ctx)
	if err != nil {
		return err
	}

	remoteRef, err := repo.Reference(n.localRemoteReferenceName(), true)
	if err != nil {
		return err
	}

	localRef, err := repo.Reference(plumbing.ReferenceName(n.LocalReference), true)
	if err != nil {
		return err
	}

	n.IsBehind, err = gitutil.HasAncestor(ctx, repo, remoteRef.Hash(), localRef.Hash())
	if err != nil {
		return err
	}

	n.IsAhead, err = gitutil.HasAncestor(ctx, repo, localRef.Hash(), remoteRef.Hash())
	if err != nil {
		return err
	}

	return nil
}

// iterateCommits creates an iterator for the commits of a reference.
func (n *Project) iterateCommits(
	ctx context.Context,
	refName plumbing.ReferenceName,
) (object.CommitIter, error) {
	repo, err := n.openRepository(ctx)
	if err != nil {
		return nil, nil
	}

	ref, err := repo.Reference(refName, true)
	if err != nil {
		return nil, err
	}

	return repo.Log(&git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
	})
}

// currentBranch returns the current branch of the repository.
// It returns nil if there isn't one.
func (n *Project) currentBranch(ctx context.Context) (*config.Branch, error) {
	if !n.IsCloned(ctx) {
		return nil, nil
	}

	repo, err := n.openRepository(ctx)
	if err != nil {
		return nil, err
	}

	return gitutil.CurrentBranch(repo, plumbing.ReferenceName(n.Reference))
}

// localRemoteReferenceName returns the name of the local reference that points to the remote.
func (n *Project) localRemoteReferenceName() plumbing.ReferenceName {
	parts := strings.Split(n.RemoteReference, "/")
	name := strings.Join(parts[2:], "/")
	return plumbing.NewRemoteReferenceName("origin", name)
}

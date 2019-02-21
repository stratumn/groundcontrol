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
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"groundcontrol/gitutil"
	"groundcontrol/relay"
	"groundcontrol/util"
)

// Project represents a project in the app.
type Project struct {
	ID               string   `json:"id"`
	Slug             string   `json:"slug"`
	Repository       string   `json:"repository"`
	Reference        string   `json:"reference"`
	RemoteReference  string   `json:"remoteReference"`
	LocalReference   string   `json:"localReference"`
	Description      *string  `json:"description"`
	WorkspaceID      string   `json:"workspaceId"`
	RemoteCommitIDs  []string `json:"remoteCommitIds"`
	LocalCommitIDs   []string `json:"localCommitIds"`
	Tasks            []Task   `json:"projects"`
	IsLoadingCommits bool     `json:"isLoadingCommits"`
	IsCloning        bool     `json:"isCloning"`
	IsPulling        bool     `json:"isPulling"`
	IsBehind         bool     `json:"isBehind"`
	IsAhead          bool     `json:"isAhead"`
}

// IsNode tells gqlgen that it implements Node.
func (Project) IsNode() {}

// Workspace returns the workspace associated with the project.
func (p Project) Workspace(ctx context.Context) Workspace {
	return MustLoadWorkspace(ctx, p.WorkspaceID)
}

// IsCloned checks if the project is cloned.
func (p Project) IsCloned(ctx context.Context) bool {
	return util.FileExists(p.Path(ctx))
}

// IsCached checks if the project is cached.
func (p Project) IsCached(ctx context.Context) bool {
	return util.FileExists(p.CachePath(ctx))
}

// RemoteCommits returns paginated remote commits.
func (p Project) RemoteCommits(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (CommitConnection, error) {
	return PaginateCommitIDSlice(
		ctx,
		p.RemoteCommitIDs,
		after,
		before,
		first,
		last,
	)
}

// LocalCommits returns paginated local commits.
func (p Project) LocalCommits(
	ctx context.Context,
	after *string,
	before *string,
	first *int,
	last *int,
) (CommitConnection, error) {
	return PaginateCommitIDSlice(
		ctx,
		p.LocalCommitIDs,
		after,
		before,
		first,
		last,
	)
}

// ReferenceShort returns the short name of the reference.
func (p Project) ReferenceShort() string {
	return plumbing.ReferenceName(p.Reference).Short()
}

// RemoteReferenceShort returns the short name of the remote reference.
func (p Project) RemoteReferenceShort() string {
	return plumbing.ReferenceName(p.RemoteReference).Short()
}

// LocalReferenceShort returns the short name of the local reference.
func (p Project) LocalReferenceShort() string {
	return plumbing.ReferenceName(p.LocalReference).Short()
}

// Path returns the path to the project.
func (p Project) Path(ctx context.Context) string {
	modelCtx := GetModelContext(ctx)
	return modelCtx.GetProjectPath(p.Workspace(ctx).Slug, p.Slug)
}

// CachePath returns the path to the project.
func (p Project) CachePath(ctx context.Context) string {
	modelCtx := GetModelContext(ctx)
	return modelCtx.GetProjectCachePath(p.Workspace(ctx).Slug, p.Slug)
}

// Clone clones and upserts the project.
func (p *Project) Clone(ctx context.Context) error {
	p.IsCloning = true

	defer func() {
		p.IsCloning = false
		p.MustStore(ctx)
	}()

	path := p.Path(ctx)
	options := &git.CloneOptions{
		URL:           p.Repository,
		ReferenceName: plumbing.ReferenceName(p.RemoteReference),
	}

	if _, err := git.PlainCloneContext(ctx, path, false, options); err != nil {
		return err
	}

	return p.Update(ctx)
}

// Pull pulls and upserts the project.
func (p *Project) Pull(ctx context.Context) error {
	p.IsPulling = true

	defer func() {
		p.IsPulling = false
		p.MustStore(ctx)
	}()

	repo, err := p.openRepository(ctx)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	options := &git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.ReferenceName(p.RemoteReference),
	}

	err = worktree.PullContext(ctx, options)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	if err != nil {
		return err
	}

	return p.Update(ctx)
}

// Update loads the latest commits and upserts the project.
func (p *Project) Update(ctx context.Context) error {
	p.IsLoadingCommits = true

	defer func() {
		p.IsLoadingCommits = false
		p.MustStore(ctx)
	}()

	if !p.IsCloned(ctx) && !p.IsCached(ctx) {
		if err := p.cloneCache(ctx); err != nil {
			return err
		}
	} else {
		if err := p.fetch(ctx); err != nil {
			return nil
		}
	}

	if err := p.updateReferences(ctx); err != nil {
		return err
	}

	remoteCommitIDs, err := p.loadCommits(ctx, p.localRemoteReferenceName())
	if err != nil {
		return err
	}

	localCommitIDs, err := p.loadCommits(ctx, plumbing.ReferenceName(p.LocalReference))
	if err != nil {
		return err
	}

	if len(remoteCommitIDs) > 0 {
		p.RemoteCommitIDs = remoteCommitIDs
	}

	if len(localCommitIDs) > 0 {
		p.LocalCommitIDs = localCommitIDs
	}

	return p.updateStatus(ctx)
}

// openRepository opens the repository of the project.
// If the project isn't cloned it will be a bare repository.
// It returns nil if the project isn't cloned or cached.
func (p Project) openRepository(ctx context.Context) (*git.Repository, error) {
	if p.IsCloned(ctx) {
		return git.PlainOpen(p.Path(ctx))
	}

	if p.IsCached(ctx) {
		return git.PlainOpen(p.CachePath(ctx))
	}

	return nil, nil
}

// cloneCache bare clones the repository into the cache.
func (p *Project) cloneCache(ctx context.Context) error {
	cachePath := p.CachePath(ctx)
	options := &git.CloneOptions{
		URL:           p.Repository,
		ReferenceName: plumbing.ReferenceName(p.RemoteReference),
	}

	_, err := git.PlainCloneContext(ctx, cachePath, true, options)
	return err
}

// fetch fetches either the cloned or the cached repository.
func (p *Project) fetch(ctx context.Context) error {
	repo, err := p.openRepository(ctx)
	if err != nil {
		return err
	}

	options := &git.FetchOptions{
		Force: !p.IsCloned(ctx),
	}

	err = repo.FetchContext(ctx, options)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	return nil
}

// loadCommits loads the commits of a reference.
func (p *Project) loadCommits(
	ctx context.Context,
	refName plumbing.ReferenceName,
) ([]string, error) {
	var commitIDs []string

	iter, err := p.iterateCommits(ctx, refName)
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
func (p *Project) updateReferences(ctx context.Context) error {
	branch, err := p.currentBranch(ctx)
	if err != nil {
		return err
	}

	if branch != nil {
		p.RemoteReference = branch.Merge.String()
		p.LocalReference = plumbing.NewBranchReferenceName(branch.Name).String()

	} else {
		p.RemoteReference = p.Reference
		p.LocalReference = p.Reference
	}

	return nil
}

// updateStatus updates the status of the project according to the local branch.
func (p *Project) updateStatus(ctx context.Context) error {
	repo, err := p.openRepository(ctx)
	if err != nil {
		return err
	}

	remoteRef, err := repo.Reference(p.localRemoteReferenceName(), true)
	if err != nil {
		return err
	}

	localRef, err := repo.Reference(plumbing.ReferenceName(p.LocalReference), true)
	if err != nil {
		return err
	}

	p.IsBehind, err = gitutil.HasAncestor(ctx, repo, remoteRef.Hash(), localRef.Hash())
	if err != nil {
		return err
	}

	p.IsAhead, err = gitutil.HasAncestor(ctx, repo, localRef.Hash(), remoteRef.Hash())
	if err != nil {
		return err
	}

	return nil
}

// iterateCommits creates an iterator for the commits of a reference.
func (p *Project) iterateCommits(
	ctx context.Context,
	refName plumbing.ReferenceName,
) (object.CommitIter, error) {
	repo, err := p.openRepository(ctx)
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
func (p Project) currentBranch(ctx context.Context) (*config.Branch, error) {
	if !p.IsCloned(ctx) {
		return nil, nil
	}

	repo, err := p.openRepository(ctx)
	if err != nil {
		return nil, err
	}

	return gitutil.CurrentBranch(repo, plumbing.ReferenceName(p.Reference))
}

// localRemoteReferenceName returns the name of the local reference that points to the remote.
func (p Project) localRemoteReferenceName() plumbing.ReferenceName {
	parts := strings.Split(p.RemoteReference, "/")
	name := strings.Join(parts[2:], "/")
	return plumbing.NewRemoteReferenceName("origin", name)
}

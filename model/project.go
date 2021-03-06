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
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"groundcontrol/appcontext"
	"groundcontrol/gitutil"
	"groundcontrol/util"
)

// String is a string representation for the type instance.
func (n *Project) String() string {
	return filepath.Base(n.Repository)
}

// LongString is a long string representation for the type instance.
func (n *Project) LongString(ctx context.Context) string {
	return fmt.Sprintf("%s » %s", n.Workspace(ctx), n)
}

// ReferenceShort is the short name of the Reference.
func (n *Project) ReferenceShort() string {
	return plumbing.ReferenceName(n.Reference).Short()
}

// RemoteReferenceShort is the short name of the RemoteReference.
func (n *Project) RemoteReferenceShort() string {
	return plumbing.ReferenceName(n.RemoteReference).Short()
}

// LocalReferenceShort is the short name of the LocalReference.
func (n *Project) LocalReferenceShort() string {
	return plumbing.ReferenceName(n.LocalReference).Short()
}

// Path is the path to the Project.
func (n *Project) Path(ctx context.Context) string {
	appCtx := appcontext.Get(ctx)
	return appCtx.GetProjectPath(n.Workspace(ctx).Slug, n.Slug)
}

// ShortPath is the path to the Project relative to the home directory.
func (n *Project) ShortPath(ctx context.Context) (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Rel(home, n.Path(ctx))
}

// IsCloned indicates whether the repository is cloned.
func (n *Project) IsCloned(ctx context.Context) bool {
	return util.FileExists(n.Path(ctx))
}

// IsCached checks if the Project is cached.
func (n *Project) IsCached(ctx context.Context) bool {
	return util.FileExists(n.CachePath(ctx))
}

// CachePath returns the path to the Project's cache.
func (n *Project) CachePath(ctx context.Context) string {
	appCtx := appcontext.Get(ctx)
	return appCtx.GetProjectCachePath(n.Workspace(ctx).Slug, n.Slug)
}

// Clone clones and store the Project.
func (n *Project) Clone(ctx context.Context) error {
	defer func() {
		n.IsCloning = false
		n.MustStore(ctx)
	}()
	n.IsCloning = true
	n.MustStore(ctx)

	path := n.Path(ctx)
	refName := plumbing.ReferenceName(n.RemoteReference)
	opts := git.CloneOptions{URL: n.Repository, ReferenceName: refName}
	if _, err := git.PlainCloneContext(ctx, path, false, &opts); err != nil {
		return err
	}
	return n.Sync(ctx)
}

// Pull pulls and stores the Project.
func (n *Project) Pull(ctx context.Context) error {
	defer func() {
		n.IsPulling = false
		n.MustStore(ctx)
	}()
	n.IsPulling = true
	n.MustStore(ctx)

	repo, err := n.openRepository(ctx)
	if err != nil {
		return err
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}
	refName := plumbing.ReferenceName(n.RemoteReference)
	opts := git.PullOptions{RemoteName: "origin", ReferenceName: refName}
	err = worktree.PullContext(ctx, &opts)
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	if err != nil {
		return err
	}
	return n.Sync(ctx)
}

// Sync syncs the Project with Git.
func (n *Project) Sync(ctx context.Context) error {
	defer func() {
		n.IsSyncing = false
		n.MustStore(ctx)
	}()
	n.IsSyncing = true
	n.MustStore(ctx)

	if err := n.syncReferences(ctx); err != nil {
		return err
	}
	if err := n.fetchOrClone(ctx); err != nil {
		return err
	}
	remoteCommitsIDs, err := n.loadCommits(ctx, n.localRemoteReferenceName())
	if err != nil {
		return err
	}
	n.RemoteCommitsIDs = remoteCommitsIDs
	localCommitsIDs, err := n.loadCommits(ctx, plumbing.ReferenceName(n.LocalReference))
	if err != nil {
		return err
	}
	n.LocalCommitsIDs = localCommitsIDs
	return n.syncStatus(ctx)
}

// EnsureCloned guarantees the Project to be cloned by the time it returns.
func (n *Project) EnsureCloned(ctx context.Context) error {
	subs := appcontext.Get(ctx).Subs
	lastMsgID := subs.LastMessageID()
	return MustLockProjectE(ctx, n.ID, func(node *Project) error {
		*n = *node
		if n.IsCloning {
			return n.waitTillCloned(ctx, lastMsgID)
		}
		if n.IsCloned(ctx) {
			return nil
		}
		return n.Clone(ctx)
	})
}

// waitTillCloned waits for the Project to finish cloning.
func (n *Project) waitTillCloned(ctx context.Context, lastMsgID uint64) error {
	subsCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	subs := appcontext.Get(ctx).Subs
	subs.Subscribe(subsCtx, MessageTypeProjectStored, lastMsgID, func(msg interface{}) {
		node := msg.(*Project)
		if node.ID != n.ID {
			return
		}
		*n = *node
		if !n.IsCloning {
			cancel()
		}
	})
	<-subsCtx.Done()
	if !n.IsCloned(ctx) {
		return ErrClone
	}
	return nil
}

// fetchOrClone fetches the repo if cloned or cached, otherwise it clones it.
func (n *Project) fetchOrClone(ctx context.Context) error {
	if !n.IsCloned(ctx) && !n.IsCached(ctx) {
		return n.cloneCache(ctx)
	}
	return n.fetch(ctx)
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
	refName := plumbing.ReferenceName(n.RemoteReference)
	opts := git.CloneOptions{URL: n.Repository, ReferenceName: refName}
	_, err := git.PlainCloneContext(ctx, cachePath, true, &opts)
	return err
}

// fetch fetches either the cloned or the cached repository.
func (n *Project) fetch(ctx context.Context) error {
	repo, err := n.openRepository(ctx)
	if err != nil {
		return err
	}
	opts := git.FetchOptions{Force: !n.IsCloned(ctx)}
	err = repo.FetchContext(ctx, &opts)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}
	return nil
}

// loadCommits loads the commits of a reference.
func (n *Project) loadCommits(ctx context.Context, refName plumbing.ReferenceName) ([]string, error) {
	iter, err := n.iterateCommits(ctx, refName)
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var commitIDs []string
	return commitIDs, iter.ForEach(func(c *object.Commit) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		commit := NewCommitFromGit(c)
		commit.MustStore(ctx)
		commitIDs = append(commitIDs, commit.ID)
		return nil
	})
}

// syncReferences sets the local and remote references according to the current branch.
func (n *Project) syncReferences(ctx context.Context) error {
	branch, err := n.currentBranch(ctx)
	if err != nil {
		return err
	}
	if branch != nil {
		n.RemoteReference = branch.Merge.String()
		n.LocalReference = plumbing.NewBranchReferenceName(branch.Name).String()
		return nil
	}
	n.RemoteReference = n.Reference
	n.LocalReference = n.Reference
	return nil
}

// syncStatus syncs the status of project according to the local branch.
func (n *Project) syncStatus(ctx context.Context) error {
	if !n.IsCloned(ctx) {
		n.IsBehind = false
		n.IsAhead = false
		n.IsClean = true
		return nil
	}
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
	n.IsClean, err = n.checkIfClean(ctx)
	return err
}

// checkIfClean checks if there are uncommited changes.
// We currently call git because go-git can't do it efficiently.
func (n *Project) checkIfClean(ctx context.Context) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = n.Path(ctx)
	out, err := cmd.CombinedOutput()
	return len(out) < 1, err
}

// iterateCommits creates an iterator for the commits of a reference.
func (n *Project) iterateCommits(ctx context.Context, refName plumbing.ReferenceName) (object.CommitIter, error) {
	repo, err := n.openRepository(ctx)
	if err != nil {
		return nil, nil
	}
	ref, err := repo.Reference(refName, true)
	if err != nil {
		return nil, err
	}
	opts := git.LogOptions{From: ref.Hash(), Order: git.LogOrderCommitterTime}
	return repo.Log(&opts)
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

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
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/relay"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

// ProjectCachePathGetter is a function that returns the path to a project's cache.
type ProjectCachePathGetter func(workspaceSlug, repo, branch string) string

// LoadCommits loads the commits of a project from a remote repo and updates the project.
func LoadCommits(
	nodes *models.NodeManager,
	jobs *models.JobManager,
	subs *pubsub.PubSub,
	getProjectPath models.ProjectPathGetter,
	getProjectCachePath ProjectCachePathGetter,
	projectID string,
	priority models.JobPriority,
) (string, error) {
	var (
		projectError error
		workspaceID  string
	)

	err := nodes.LockProject(projectID, func(project models.Project) {
		if project.IsLoadingCommits {
			projectError = ErrDuplicate
			return
		}

		workspaceID = project.WorkspaceID
		project.IsLoadingCommits = true
		nodes.MustStoreProject(project)
	})
	if err != nil {
		return "", err
	}
	if projectError != nil {
		return "", projectError
	}

	subs.Publish(models.ProjectUpdated, projectID)
	subs.Publish(models.WorkspaceUpdated, workspaceID)

	jobID := jobs.Add(
		LoadCommitsJob,
		projectID,
		priority,
		func(ctx context.Context) error {
			return doLoadCommits(
				ctx,
				nodes,
				subs,
				getProjectPath,
				getProjectCachePath,
				projectID,
				workspaceID,
			)
		},
	)

	return jobID, nil
}

func doLoadCommits(
	ctx context.Context,
	nodes *models.NodeManager,
	subs *pubsub.PubSub,
	getProjectPath models.ProjectPathGetter,
	getProjectCachePath ProjectCachePathGetter,
	projectID string,
	workspaceID string,
) error {
	var (
		commitIDs []string
		err       error
	)

	defer func() {
		nodes.MustLockProject(projectID, func(project models.Project) {
			if len(commitIDs) > 0 {
				project.CommitIDs = commitIDs
			}

			project.IsLoadingCommits = false
			nodes.MustStoreProject(project)
		})

		subs.Publish(models.ProjectUpdated, projectID)
		subs.Publish(models.WorkspaceUpdated, workspaceID)
	}()

	project := nodes.MustLoadProject(projectID)

	repo, err := cloneOrFetch(
		ctx,
		nodes,
		getProjectPath,
		getProjectCachePath,
		projectID,
	)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	refName := plumbing.NewRemoteReferenceName("origin", project.Branch)
	ref, err := repo.Reference(refName, true)
	if err != nil {
		return err
	}

	commitIDs, err = getCommits(ctx, nodes, repo, ref)
	if err != nil {
		return err
	}

	if project.IsCloned(nodes, getProjectPath) {
		return updateStatus(ctx, nodes, repo, projectID)
	}

	return nil
}

func cloneOrFetch(
	ctx context.Context,
	nodes *models.NodeManager,
	getProjectPath models.ProjectPathGetter,
	getProjectCachePath ProjectCachePathGetter,
	projectID string,
) (repo *git.Repository, err error) {
	project := nodes.MustLoadProject(projectID)
	workspace := project.Workspace(nodes)
	projectDir := getProjectPath(workspace.Slug, project.Repository, project.Branch)
	cacheDir := getProjectCachePath(workspace.Slug, project.Repository, project.Branch)
	force := false

	var refSpec []config.RefSpec

	if project.IsCloned(nodes, getProjectPath) {
		repo, err = git.PlainOpen(projectDir)
	} else if exists(cacheDir) {
		repo, err = git.PlainOpen(cacheDir)
		refSpec = []config.RefSpec{"+refs/heads/*:refs/heads/*"}
		force = true
	}

	if err != nil {
		return
	}

	if repo != nil {
		err = repo.FetchContext(
			ctx,
			&git.FetchOptions{
				RefSpecs: refSpec,
				Force:    force,
			},
		)
	} else {
		repo, err = git.PlainCloneContext(
			ctx,
			cacheDir, true,
			&git.CloneOptions{
				URL:           project.Repository,
				ReferenceName: plumbing.NewBranchReferenceName(project.Branch),
			},
		)
	}

	return
}

func getCommits(
	ctx context.Context,
	nodes *models.NodeManager,
	repo *git.Repository,
	ref *plumbing.Reference,
) ([]string, error) {
	var commitIDs []string

	iter, err := repo.Log(&git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, err
	}

	return commitIDs, iter.ForEach(func(c *object.Commit) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		commit := models.Commit{
			ID:       relay.EncodeID(models.NodeTypeCommit, c.Hash.String()),
			Hash:     models.Hash(c.Hash.String()),
			Headline: strings.Split(c.Message, "\n")[0],
			Message:  c.Message,
			Author:   c.Author.Name,
			Date:     models.DateTime(c.Author.When),
		}

		nodes.MustStoreCommit(commit)
		commitIDs = append(commitIDs, commit.ID)

		return nil
	})
}

func updateStatus(
	ctx context.Context,
	nodes *models.NodeManager,
	repo *git.Repository,
	projectID string,
) (projectError error) {
	nodes.MustLockProject(projectID, func(project models.Project) {
		refName := plumbing.NewBranchReferenceName(project.Branch)
		ref, err := repo.Reference(refName, true)
		if err != nil {
			projectError = err
			return
		}

		remoteRefName := plumbing.NewRemoteReferenceName("origin", project.Branch)
		remoteRef, err := repo.Reference(remoteRefName, true)
		if err != nil {
			projectError = err
			return
		}

		iter, err := repo.Log(&git.LogOptions{
			From:  ref.Hash(),
			Order: git.LogOrderCommitterTime,
		})
		if err != nil {
			projectError = err
			return
		}

		project.IsBehind = true
		project.IsAhead = false
		remoteHash := remoteRef.Hash()
		last := true

		projectError = iter.ForEach(func(commit *object.Commit) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if bytes.Compare(remoteHash[:], commit.Hash[:]) == 0 {
				project.IsBehind = false
				project.IsAhead = !last
				return storer.ErrStop
			}

			last = false
			return nil
		})
		if projectError != nil {
			return
		}

		nodes.MustStoreProject(project)
	})

	return
}

func exists(directory string) bool {
	_, err := os.Stat(directory)
	return !os.IsNotExist(err)
}

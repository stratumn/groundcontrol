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

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"

	"groundcontrol/models"
	"groundcontrol/relay"
)

// LoadCommits loads the commits of a project from a remote repo and updates the project.
func LoadCommits(ctx context.Context, projectID string, priority models.JobPriority) (string, error) {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	subs := modelCtx.Subs
	workspaceID := ""

	err := nodes.LockProjectE(projectID, func(project models.Project) error {
		if project.IsLoadingCommits {
			return ErrDuplicate
		}

		workspaceID = project.WorkspaceID
		project.IsLoadingCommits = true
		nodes.MustStoreProject(project)

		return nil
	})
	if err != nil {
		return "", err
	}

	subs.Publish(models.ProjectUpserted, projectID)
	subs.Publish(models.WorkspaceUpserted, workspaceID)

	jobID := modelCtx.Jobs.Add(
		models.GetModelContext(ctx),
		LoadCommitsJob,
		projectID,
		priority,
		func(ctx context.Context) error {
			return doLoadCommits(ctx, projectID, workspaceID)
		},
	)

	return jobID, nil
}

func doLoadCommits(ctx context.Context, projectID string, workspaceID string) error {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	subs := modelCtx.Subs
	commitIDs := []string(nil)

	defer func() {
		nodes.MustLockProject(projectID, func(project models.Project) {
			if len(commitIDs) > 0 {
				project.CommitIDs = commitIDs
			}

			project.IsLoadingCommits = false
			nodes.MustStoreProject(project)
		})

		subs.Publish(models.ProjectUpserted, projectID)
		subs.Publish(models.WorkspaceUpserted, workspaceID)
	}()

	project := nodes.MustLoadProject(projectID)

	repo, err := cloneOrFetch(ctx, projectID)
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return err
	}

	refName := plumbing.NewRemoteReferenceName("origin", project.Branch)
	ref, err := repo.Reference(refName, true)
	if err != nil {
		return err
	}

	commitIDs, err = getCommits(ctx, repo, ref)
	if err != nil {
		return err
	}

	if project.IsCloned(ctx) {
		return updateStatus(ctx, repo, projectID)
	}

	return nil
}

func cloneOrFetch(ctx context.Context, projectID string) (repo *git.Repository, err error) {
	modelCtx := models.GetModelContext(ctx)
	project := modelCtx.Nodes.MustLoadProject(projectID)
	workspace := project.Workspace(ctx)
	projectDir := modelCtx.GetProjectPath(workspace.Slug, project.Repository, project.Branch)
	cacheDir := modelCtx.GetProjectCachePath(workspace.Slug, project.Repository, project.Branch)
	force := false

	if project.IsCloned(ctx) {
		repo, err = git.PlainOpen(projectDir)
	} else if exists(cacheDir) {
		repo, err = git.PlainOpen(cacheDir)
		force = true
	}

	if err != nil {
		return
	}

	if repo != nil {
		err = repo.FetchContext(
			ctx,
			&git.FetchOptions{
				Force: force,
			},
		)
	} else {
		repo, err = git.PlainCloneContext(
			ctx,
			cacheDir,
			true,
			&git.CloneOptions{
				URL:           project.Repository,
				ReferenceName: plumbing.NewBranchReferenceName(project.Branch),
			},
		)
	}

	return
}

func getCommits(ctx context.Context, repo *git.Repository, ref *plumbing.Reference) ([]string, error) {
	var commitIDs []string

	modelCtx := models.GetModelContext(ctx)

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

		modelCtx.Nodes.MustStoreCommit(commit)
		commitIDs = append(commitIDs, commit.ID)

		return nil
	})
}

func updateStatus(ctx context.Context, repo *git.Repository, projectID string) (projectError error) {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes

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

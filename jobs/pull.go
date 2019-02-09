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
	"strings"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"

	"groundcontrol/models"
	"groundcontrol/relay"
)

// Pull clones a remote repository locally and updates the project.
func Pull(ctx context.Context, projectID string, priority models.JobPriority) (string, error) {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	subs := modelCtx.Subs
	workspaceID := ""

	err := nodes.LockProjectE(projectID, func(project models.Project) error {
		if !project.IsCloned(ctx) {
			return ErrNotCloned
		}

		if project.IsPulling {
			return ErrDuplicate
		}

		workspaceID = project.WorkspaceID
		project.IsPulling = true
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
		PullJob,
		projectID,
		priority,
		func(ctx context.Context) error {
			return doPull(ctx, projectID, workspaceID)
		},
	)

	return jobID, nil
}

func doPull(ctx context.Context, projectID string, workspaceID string) error {
	modelCtx := models.GetModelContext(ctx)
	nodes := modelCtx.Nodes
	subs := modelCtx.Subs

	defer func() {
		nodes.MustLockProject(projectID, func(project models.Project) {
			project.IsPulling = false
			nodes.MustStoreProject(project)
		})

		subs.Publish(models.ProjectUpserted, projectID)
		subs.Publish(models.WorkspaceUpserted, workspaceID)
	}()

	project := nodes.MustLoadProject(projectID)
	workspace := project.Workspace(ctx)
	directory := modelCtx.GetProjectPath(workspace.Slug, project.Repository, project.Branch)

	repo, err := git.PlainOpen(directory)
	if err != nil {
		return err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = worktree.PullContext(ctx, &git.PullOptions{RemoteName: "origin"})
	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	if err != nil {
		return err
	}

	refName := plumbing.NewBranchReferenceName(project.Branch)
	ref, err := repo.Reference(refName, true)
	if err != nil {
		return err
	}

	iter, err := repo.Log(&git.LogOptions{
		From:  ref.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return err
	}

	var commitIDs []string

	err = iter.ForEach(func(c *object.Commit) error {
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
	if err != nil {
		return err
	}

	remoteRefName := plumbing.NewRemoteReferenceName("origin", project.Branch)
	remoteRef, err := repo.Reference(remoteRefName, true)
	if err != nil {
		return err
	}

	hash := ref.Hash()
	remoteHash := remoteRef.Hash()

	nodes.MustLockProject(projectID, func(project models.Project) {
		project.CommitIDs = commitIDs
		project.IsBehind = false
		project.IsAhead = false

		if bytes.Compare(hash[:], remoteHash[:]) != 0 {
			project.IsAhead = true
		}

		nodes.MustStoreProject(project)
	})

	subs.Publish(models.ProjectUpserted, projectID)
	subs.Publish(models.WorkspaceUpserted, workspaceID)

	return nil
}

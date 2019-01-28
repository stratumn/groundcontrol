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
	"strings"

	"github.com/stratumn/groundcontrol/date"
	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/relay"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// LoadCommits loads the commits of a project from a remote repo.
func LoadCommits(
	nodes *models.NodeManager,
	jobs *models.JobManager,
	subs *pubsub.PubSub,
	projectID string,
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

	jobID := jobs.Add(LoadCommitsJob, projectID, func() error {
		return doLoadCommits(
			nodes,
			subs,
			projectID,
			workspaceID,
		)
	})

	return jobID, nil
}

func doLoadCommits(
	nodes *models.NodeManager,
	subs *pubsub.PubSub,
	projectID string,
	workspaceID string,
) error {
	var commitIDs []string

	defer func() {
		nodes.MustLockProject(projectID, func(project models.Project) {
			project.CommitIDs = commitIDs
			project.IsLoadingCommits = false
			nodes.MustStoreProject(project)
		})

		subs.Publish(models.ProjectUpdated, projectID)
		subs.Publish(models.WorkspaceUpdated, workspaceID)
	}()

	project := nodes.MustLoadProject(projectID)
	refName := plumbing.NewBranchReferenceName(project.Branch)

	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           project.Repository,
		ReferenceName: refName,
	})
	if err != nil {
		return err
	}

	ref, err := repo.Reference(refName, true)
	if err != nil {
		return err
	}

	iter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}

	return iter.ForEach(func(c *object.Commit) error {
		commit := models.Commit{
			ID:       relay.EncodeID(models.NodeTypeCommit, c.Hash.String()),
			SHA:      c.Hash.String(),
			Headline: strings.Split(c.Message, "\n")[0],
			Message:  c.Message,
			Author:   c.Author.Name,
			Date:     c.Author.When.Format(date.DateFormat),
		}

		nodes.MustStoreCommit(commit)
		commitIDs = append(commitIDs, commit.ID)

		return nil
	})
}

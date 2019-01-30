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
	"os"
	"strings"

	"github.com/stratumn/groundcontrol/date"
	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/relay"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// ProjectCachePathGetter is a function that returns the path to a project's cache.
type ProjectCachePathGetter func(workspaceSlug, repo, branch string) string

// LoadCommits loads the commits of a project from a remote repo.
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
		func() error {
			return doLoadCommits(
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
		nodes,
		getProjectPath,
		getProjectCachePath,
		projectID,
	)
	if err == git.NoErrAlreadyUpToDate {
		if len(project.CommitIDs) > 0 {
			return nil
		}
	} else if err != nil {
		return err
	}

	refName := plumbing.NewRemoteReferenceName("origin", project.Branch)

	ref, err := repo.Reference(refName, true)
	if err != nil {
		return err
	}

	commitIDs, err = getCommits(nodes, repo, ref)

	return err
}

func cloneOrFetch(
	nodes *models.NodeManager,
	getProjectPath models.ProjectPathGetter,
	getProjectCachePath ProjectCachePathGetter,
	projectID string,
) (repo *git.Repository, err error) {
	project := nodes.MustLoadProject(projectID)
	workspace := project.Workspace(nodes)
	projectDir := getProjectPath(workspace.Slug, project.Repository, project.Branch)
	cacheDir := getProjectCachePath(workspace.Slug, project.Repository, project.Branch)

	if project.IsCloned(nodes, getProjectPath) {
		repo, err = git.PlainOpen(projectDir)
	} else if exists(cacheDir) {
		repo, err = git.PlainOpen(cacheDir)
	}

	if err != nil {
		return
	}

	if repo != nil {
		err = repo.Fetch(&git.FetchOptions{
			RefSpecs: []config.RefSpec{},
		})
	} else {
		repo, err = git.PlainClone(cacheDir, true, &git.CloneOptions{
			URL:           project.Repository,
			ReferenceName: plumbing.NewBranchReferenceName(project.Branch),
		})
	}

	return
}

func getCommits(
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

func exists(directory string) bool {
	_, err := os.Stat(directory)
	return !os.IsNotExist(err)
}

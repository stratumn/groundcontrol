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
	"container/list"
	"errors"
	"os"
	"strings"
	"sync/atomic"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/stratumn/groundcontrol/date"
	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/relay"
)

// Job names.
const (
	LoadCommitsJob = "Load Commits"
	CloneJob       = "Clone"
)

// Message types.
const (
	ProjectUpdated = "PROJECT_UPDATED" // Go type *Project
)

var commitPaginator = relay.Paginator{
	GetID: func(node interface{}) string {
		return node.(*Commit).ID
	},
}

// Errors.
var (
	ErrCloning = errors.New("project is being cloned")
	ErrCloned  = errors.New("project is  cloned")
)

// ProjectPathGetter is a function that returns the path to a project.
type ProjectPathGetter func(workspaceSlug, repo, branch string) string

// Project represents a project in the app.
type Project struct {
	ID          string     `json:"id"`
	Repository  string     `json:"repository"`
	Branch      string     `json:"branch"`
	Description *string    `json:"description"`
	Workspace   *Workspace `json:"workspace"`

	commitList *list.List

	isLoadingCommits uint32
	isCloning        uint32
}

// IsNode is used by gqlgen.
func (*Project) IsNode() {}

// IsCloned checks if the project is cloned.
func (p *Project) IsCloned(getProjectPath ProjectPathGetter) bool {
	return p.isCloned(getProjectPath(p.Workspace.Slug, p.Repository, p.Branch))
}

func (p *Project) isCloned(directory string) bool {
	_, err := os.Stat(directory)
	return !os.IsNotExist(err)
}

// IsCloning returns whether the project is beining cloned.
func (p *Project) IsCloning() bool {
	return atomic.LoadUint32(&p.isCloning) == 1
}

// Commits returns paginated commits.
// If there are no commits in memory, it may create a LoadCommitJob.
func (p *Project) Commits(
	nodes *NodeManager,
	jobManager *JobManager,
	pubsub *pubsub.PubSub,
	after *string,
	before *string,
	first *int,
	last *int,
) (CommitConnection, error) {
	if p.commitList.Len() == 0 {
		p.loadCommitsJob(nodes, jobManager, pubsub)

		return CommitConnection{
			IsLoading: true,
		}, nil
	}

	connection, err := commitPaginator.Paginate(p.commitList, after, before, first, last)
	if err != nil {
		return CommitConnection{}, err
	}

	edges := make([]CommitEdge, len(connection.Edges))

	for i, v := range connection.Edges {
		edges[i] = CommitEdge{
			Node:   *v.Node.(*Commit),
			Cursor: v.Cursor,
		}
	}

	return CommitConnection{
		Edges:     edges,
		PageInfo:  connection.PageInfo,
		IsLoading: atomic.LoadUint32(&p.isLoadingCommits) == 1,
	}, nil
}

func (p *Project) loadCommitsJob(
	nodes *NodeManager,
	jobManager *JobManager,
	pubsub *pubsub.PubSub,
) {
	if !atomic.CompareAndSwapUint32(&p.isLoadingCommits, 0, 1) {
		return
	}

	pubsub.Publish(ProjectUpdated, p)
	pubsub.Publish(WorkspaceUpdated, p.Workspace)

	jobManager.Add(LoadCommitsJob, p, func() error {
		err := p.loadCommits(nodes)
		atomic.StoreUint32(&p.isLoadingCommits, 0)
		pubsub.Publish(ProjectUpdated, p)
		pubsub.Publish(WorkspaceUpdated, p.Workspace)
		return err
	})
}

func (p *Project) loadCommits(nodes *NodeManager) error {
	refName := plumbing.NewBranchReferenceName(p.Branch)

	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           p.Repository,
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
		commit := &Commit{
			ID:       c.Hash.String(),
			Headline: strings.Split(c.Message, "\n")[0],
			Message:  c.Message,
			Author:   c.Author.Name,
			Date:     c.Author.When.Format(date.DateFormat),
		}

		nodes.Store(commit.ID, commit)
		p.commitList.PushBack(commit)

		return nil
	})
}

// CloneJob add a job to clone the repo if there isn't alreay one.
func (p *Project) CloneJob(
	jobManager *JobManager,
	pubsub *pubsub.PubSub,
	getProjectPath ProjectPathGetter,
) (*Job, error) {
	if !atomic.CompareAndSwapUint32(&p.isCloning, 0, 1) {
		return nil, ErrCloning
	}

	pubsub.Publish(ProjectUpdated, p)
	pubsub.Publish(WorkspaceUpdated, p.Workspace)

	return jobManager.Add(CloneJob, p, func() error {
		err := p.clone(getProjectPath)
		atomic.StoreUint32(&p.isCloning, 0)
		pubsub.Publish(ProjectUpdated, p)
		pubsub.Publish(WorkspaceUpdated, p.Workspace)
		return err
	}), nil
}

func (p *Project) clone(getProjectPath ProjectPathGetter) error {
	directory := getProjectPath(p.Workspace.Slug, p.Repository, p.Branch)

	if p.isCloned(directory) {
		return ErrCloned
	}

	_, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:           p.Repository,
		ReferenceName: plumbing.NewBranchReferenceName(p.Branch),
	})

	return err
}

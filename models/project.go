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
	"strings"
	"sync"

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
)

// Message types.
const (
	ProjectUpdated = "PROJECT_UPDATED" // Go type *Project
)

var commitPaginator = relay.Paginator{
	GetID: func(node interface{}) string {
		return node.(Commit).ID
	},
}

// Project represents a project in the app.
type Project struct {
	ID          string     `json:"id"`
	Repository  string     `json:"repository"`
	Branch      string     `json:"branch"`
	Description *string    `json:"description"`
	IsCloning   bool       `json:"isCloning"`
	IsCloned    bool       `json:"isCloned"`
	Workspace   *Workspace `json:"workspace"`

	commitList *list.List

	commitsMu        sync.Mutex
	isLoadingCommits bool

	cloneMu   sync.Mutex
	isCloning bool
}

// IsNode is used by gqlgen.
func (*Project) IsNode() {}

// Commits returns paginated commits.
// If there are no commits in memory, it may create a LoadCommitJob.
func (p *Project) Commits(
	jobManager *JobManager,
	pubsub *pubsub.PubSub,
	after *string,
	before *string,
	first *int,
	last *int,
) (CommitConnection, error) {
	if p.commitList.Len() == 0 {
		p.loadCommitsJob(jobManager, pubsub)

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
			Node:   v.Node.(Commit),
			Cursor: v.Cursor,
		}
	}

	return CommitConnection{
		Edges:     edges,
		PageInfo:  connection.PageInfo,
		IsLoading: p.isLoadingCommits,
	}, nil
}

func (p *Project) loadCommitsJob(jobManager *JobManager, pubsub *pubsub.PubSub) {
	p.commitsMu.Lock()
	defer p.commitsMu.Unlock()

	if !p.isLoadingCommits {
		p.isLoadingCommits = true

		jobManager.Add(
			LoadCommitsJob,
			p,
			func() error {
				err := p.loadCommits()
				p.commitsMu.Lock()
				p.isLoadingCommits = false
				p.commitsMu.Unlock()
				pubsub.Publish(ProjectUpdated, p)
				pubsub.Publish(WorkspaceUpdated, p.Workspace)
				return err
			},
		)
	}
}

func (p *Project) loadCommits() error {
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:           p.Repository,
		ReferenceName: plumbing.NewBranchReferenceName(p.Branch),
	})
	if err != nil {
		return err
	}

	ref, err := repo.Head()
	if err != nil {
		return err
	}

	iter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return err
	}

	return iter.ForEach(func(c *object.Commit) error {
		p.commitList.PushBack(Commit{
			ID:       c.Hash.String(),
			Headline: strings.Split(c.Message, "\n")[0],
			Message:  c.Message,
			Author:   c.Author.Name,
			Date:     c.Author.When.Format(date.DateFormat),
		})

		return nil
	})
}

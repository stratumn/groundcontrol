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
	"os"
)

// ProjectPathGetter is a function that returns the path to a project.
type ProjectPathGetter func(workspaceSlug, repo, branch string) string

// Project represents a project in the app.
type Project struct {
	ID               string   `json:"id"`
	Repository       string   `json:"repository"`
	Branch           string   `json:"branch"`
	Description      *string  `json:"description"`
	WorkspaceID      string   `json:"workspaceID"`
	CommitIDs        []string `json:"commitIDs"`
	Tasks            []Task   `json:"projects"`
	IsLoadingCommits bool     `json:"isLoadingCommits"`
	IsCloning        bool     `json:"isCloning"`
	IsPulling        bool     `json:"isPulling"`
}

// IsNode tells gqlgen that it implements Node.
func (Project) IsNode() {}

// Workspace returns the workspace associated with the project.
func (p Project) Workspace(nodes *NodeManager) Workspace {
	return nodes.MustLoadWorkspace(p.WorkspaceID)
}

// IsCloned checks if the project is cloned.
func (p Project) IsCloned(
	nodes *NodeManager,
	getProjectPath ProjectPathGetter,
) bool {
	directory := getProjectPath(p.Workspace(nodes).Slug, p.Repository, p.Branch)
	return p.isCloned(directory)
}

func (p Project) isCloned(directory string) bool {
	_, err := os.Stat(directory)
	return !os.IsNotExist(err) && !p.IsCloning
}

// Commits returns paginated commits.
// If there are no commits in memory, it may create a LoadCommitJob.
func (p Project) Commits(
	nodes *NodeManager,
	after *string,
	before *string,
	first *int,
	last *int,
) (CommitConnection, error) {
	var slice []Commit

	for _, nodeID := range p.CommitIDs {
		slice = append(slice, nodes.MustLoadCommit(nodeID))
	}

	return PaginateCommitSlice(
		slice,
		after,
		before,
		first,
		last,
	)
}

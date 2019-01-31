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

// User contains all the data of the person using the app.
type User struct {
	ID           string   `json:"id"`
	WorkspaceIDs []string `json:"workspaceIDs"`
}

// IsNode tells gqlgen that it implements Node.
func (User) IsNode() {}

// Workspaces returns the user's workspaces.
func (u User) Workspaces(nodes *NodeManager) []Workspace {
	var slice []Workspace

	for _, nodeID := range u.WorkspaceIDs {
		slice = append(slice, nodes.MustLoadWorkspace(nodeID))
	}

	return slice
}

// Workspace finds a workspace.
func (u User) Workspace(nodes *NodeManager, slug string) *Workspace {
	for _, id := range u.WorkspaceIDs {
		node := nodes.MustLoadWorkspace(id)

		if node.Slug == slug {
			return &node
		}
	}

	return nil
}

// Projects returns the user's projects.
func (u User) Projects(nodes *NodeManager) []Project {
	var slice []Project

	for _, workspace := range u.Workspaces(nodes) {
		slice = append(slice, workspace.Projects(nodes)...)
	}

	return slice
}

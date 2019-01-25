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

// Workspace represents a workspace in the app.
type Workspace struct {
	ID          string   `json:"id"`
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	ProjectIDs  []string `json:"projectIDs"`
	TaskIDs     []string `json:"taskIDs"`
	Description string   `json:"description"`
	Notes       *string  `json:"notes"`
}

// IsNode tells gqlgen that it implements Node.
func (Workspace) IsNode() {}

// Projects returns the workspace's projects.
func (w Workspace) Projects(nodes *NodeManager) []Project {
	var slice []Project

	for _, nodeID := range w.ProjectIDs {
		slice = append(slice, nodes.MustLoadProject(nodeID))
	}

	return slice
}

// Tasks returns the workspace's tasks.
func (w Workspace) Tasks(nodes *NodeManager) []Task {
	var slice []Task

	for _, nodeID := range w.TaskIDs {
		slice = append(slice, nodes.MustLoadTask(nodeID))
	}

	return slice
}

// IsCloning returns true if any of the projects is cloning.
func (w Workspace) IsCloning(nodes *NodeManager) bool {
	for _, v := range w.Projects(nodes) {
		if v.IsCloning {
			return true
		}
	}

	return false
}

// IsCloned returns true if all the projects are cloned.
func (w Workspace) IsCloned(nodes *NodeManager, getProjectPath ProjectPathGetter) bool {
	for _, v := range w.Projects(nodes) {
		if !v.IsCloned(nodes, getProjectPath) {
			return false
		}
	}

	return true
}

// IsPulling returns true if any of the projects is pulling.
func (w Workspace) IsPulling(nodes *NodeManager) bool {
	for _, v := range w.Projects(nodes) {
		if v.IsPulling {
			return true
		}
	}

	return false
}

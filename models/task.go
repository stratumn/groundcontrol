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

// Task represents a workspace task in the app.
type Task struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	StepIDs     []string `json:"stepIDs"`
	WorkspaceID string   `json:"workspace"`
}

// IsNode tells gqlgen that it implements Node.
func (Task) IsNode() {}

// Steps returns the task's steps.
func (t Task) Steps(nodes *NodeManager) []Step {
	var slice []Step

	for _, nodeID := range t.StepIDs {
		slice = append(slice, nodes.MustLoadStep(nodeID))
	}

	return slice
}

// Workspace returns the task's workspace.
func (t Task) Workspace(nodes *NodeManager) Workspace {
	return nodes.MustLoadWorkspace(t.WorkspaceID)
}

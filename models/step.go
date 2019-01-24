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

// Step represents a task step in the app.
type Step struct {
	ID         string   `json:"id"`
	ProjectIDs []string `json:"projectIDs"`
	Commands   []string `json:"commands"`
	Background bool     `json:"background"`
	TaskID     string   `json:"taskID"`
}

// IsNode tells gqlgen that it implements Node.
func (Step) IsNode() {}

// Projects returns the step's projects.
func (s Step) Projects(nodes *NodeManager) []Project {
	var slice []Project

	for _, nodeID := range s.ProjectIDs {
		slice = append(slice, nodes.MustLoadProject(nodeID))
	}

	return slice
}

// Task returns the step's taks.
func (s Step) Task(nodes *NodeManager) Task {
	return nodes.MustLoadTask(s.TaskID)
}

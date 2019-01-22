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

// Job represents a job in the app.
type Job struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt string    `json:"createdAt"`
	UpdatedAt string    `json:"updatedAt"`
	Status    JobStatus `json:"status"`
	ProjectID string    `json:"projectID"`
}

// IsNode tells gqlgen that it implements Node.
func (Job) IsNode() {}

// Project returns the project associated with the job.
func (j Job) Project(nodes *NodeManager) Project {
	return nodes.MustLoadProject(j.ProjectID)
}

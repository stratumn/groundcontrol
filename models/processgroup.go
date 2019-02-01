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

// ProcessGroup represents a ProcessGroup in the app.
type ProcessGroup struct {
	ID         string   `json:"id"`
	CreatedAt  DateTime `json:"createdAt"`
	TaskID     string   `json:"taskID"`
	ProcessIDs []string `json:"processIDs"`
}

// IsNode tells gqlgen that it implements Node.
func (ProcessGroup) IsNode() {}

// Processes returns the ProcessGroup's processes.
func (p ProcessGroup) Processes(
	nodes *NodeManager,
	after *string,
	before *string,
	first *int,
	last *int,
) (ProcessConnection, error) {
	return PaginateProcessIDSlice(nodes, p.ProcessIDs, after, before, first, last)
}

// Task returns the Task associated with the ProcessGroup.
func (p ProcessGroup) Task(nodes *NodeManager) Task {
	return nodes.MustLoadTask(p.TaskID)
}

// Status returns the status of the ProcessGroup.
func (p ProcessGroup) Status(nodes *NodeManager) ProcessStatus {
	status := ProcessStatusDone

	for _, id := range p.ProcessIDs {
		node := nodes.MustLoadProcess(id)

		if node.Status == ProcessStatusFailed {
			return ProcessStatusFailed
		}
		if node.Status == ProcessStatusRunning {
			status = ProcessStatusRunning
		}
	}

	return status
}

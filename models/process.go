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

import "context"

// Process represents a process in the app.
type Process struct {
	ID      string `json:"id"`
	Command string `json:"command"`
	// Env is the environment of the process.
	// Each entry is of the form "key=value".
	Env            []string      `json:"env"`
	ProcessGroupID string        `json:"processGroupId"`
	ProjectID      string        `json:"projectId"`
	Status         ProcessStatus `json:"status"`
}

// IsNode tells gqlgen that it implements Node.
func (Process) IsNode() {}

// ProcessGroup returns the ProcessGroup associated with the Process.
func (p Process) ProcessGroup(ctx context.Context) ProcessGroup {
	return GetModelContext(ctx).Nodes.MustLoadProcessGroup(p.ProcessGroupID)
}

// Project returns the Project associated with the Process.
func (p Process) Project(ctx context.Context) Project {
	return GetModelContext(ctx).Nodes.MustLoadProject(p.ProjectID)
}

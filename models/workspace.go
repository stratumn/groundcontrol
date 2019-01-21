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

// Message types.
const (
	WorkspaceUpdated = "WORKSPACE_UPDATED" // Go type *Workspace
)

type Workspace struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Projects    []Project `json:"projects"`
	Description string    `json:"description"`
	Notes       *string   `json:"notes"`
}

func (Workspace) IsNode() {}

func (w Workspace) IsCloning() bool {
	for _, v := range w.Projects {
		if v.IsCloning {
			return true
		}
	}

	return false
}

func (w Workspace) IsCloned() bool {
	for _, v := range w.Projects {
		if !v.IsCloned {
			return false
		}
	}

	return true
}

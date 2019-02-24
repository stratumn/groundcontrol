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

// IsCloning returns true if any of the projects is cloning.
func (n *Workspace) IsCloning(ctx context.Context) bool {
	for _, id := range n.ProjectsIDs {
		node := MustLoadProject(ctx, id)
		if node.IsCloning {
			return true
		}
	}

	return false
}

// IsCloned returns true if all the projects are cloned.
func (n *Workspace) IsCloned(ctx context.Context) bool {
	for _, id := range n.ProjectsIDs {
		node := MustLoadProject(ctx, id)
		if !node.IsCloned(ctx) {
			return false
		}
	}

	return true
}

// IsPulling returns true if any of the projects is pulling.
func (n *Workspace) IsPulling(ctx context.Context) bool {
	for _, id := range n.ProjectsIDs {
		node := MustLoadProject(ctx, id)
		if node.IsPulling {
			return true
		}
	}

	return false
}

// IsBehind returns true if any of the projects is behind origin.
func (n *Workspace) IsBehind(ctx context.Context) bool {
	for _, id := range n.ProjectsIDs {
		node := MustLoadProject(ctx, id)
		if node.IsBehind {
			return true
		}
	}

	return false
}

// IsAhead returns true if any of the projects is ahead origin.
func (n *Workspace) IsAhead(ctx context.Context) bool {
	for _, id := range n.ProjectsIDs {
		node := MustLoadProject(ctx, id)
		if node.IsAhead {
			return true
		}
	}

	return false
}

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

package model

import (
	"context"
	"sort"
	"strings"
)

// BeforeStore sorts the Workspaces by Slug before storing the Workspace.
func (n *Workspace) BeforeStore(ctx context.Context) {
	sort.Slice(n.ProjectsIDs, func(i, j int) bool {
		a := MustLoadProject(ctx, n.ProjectsIDs[i])
		b := MustLoadProject(ctx, n.ProjectsIDs[j])
		return strings.ToLower(a.Slug) < strings.ToLower(b.Slug)
	})
}

// IsCloning indicates whether any of the Projects is currently cloning.
func (n *Workspace) IsCloning(ctx context.Context) bool {
	for _, id := range n.ProjectsIDs {
		node := MustLoadProject(ctx, id)
		if node.IsCloning {
			return true
		}
	}

	return false
}

// IsCloned indicates whether all Projects are cloned.
func (n *Workspace) IsCloned(ctx context.Context) bool {
	for _, id := range n.ProjectsIDs {
		node := MustLoadProject(ctx, id)
		if !node.IsCloned(ctx) {
			return false
		}
	}

	return true
}

// IsPulling indicates whether any of the Projects is currently pulling.
func (n *Workspace) IsPulling(ctx context.Context) bool {
	for _, id := range n.ProjectsIDs {
		node := MustLoadProject(ctx, id)
		if node.IsPulling {
			return true
		}
	}

	return false
}

// IsBehind indicates whether any of the Projects is behind. See Project.
func (n *Workspace) IsBehind(ctx context.Context) bool {
	for _, id := range n.ProjectsIDs {
		node := MustLoadProject(ctx, id)
		if node.IsBehind {
			return true
		}
	}

	return false
}

// IsAhead indicates whether any of the Projects is ahead. See Project.
func (n *Workspace) IsAhead(ctx context.Context) bool {
	for _, id := range n.ProjectsIDs {
		node := MustLoadProject(ctx, id)
		if node.IsAhead {
			return true
		}
	}

	return false
}

// IsClean indicates whether all Projects are clean. See Project.
func (n *Workspace) IsClean(ctx context.Context) bool {
	for _, id := range n.ProjectsIDs {
		node := MustLoadProject(ctx, id)
		if !node.IsClean {
			return false
		}
	}

	return true
}

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

// BeforeStore sorts collections before storing the Workspace.
func (n *Workspace) BeforeStore(ctx context.Context) {
	n.SortProjects(ctx)
}

// SortProjects sorts the Projects by Slug.
func (n *Workspace) SortProjects(ctx context.Context) {
	sort.Slice(n.ProjectsIDs, func(i, j int) bool {
		a := MustLoadProject(ctx, n.ProjectsIDs[i])
		b := MustLoadProject(ctx, n.ProjectsIDs[j])
		return strings.ToLower(a.Slug) < strings.ToLower(b.Slug)
	})
}

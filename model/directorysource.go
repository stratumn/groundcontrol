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

import "context"

// Update loads the workspaces and upserts the source.
func (n *DirectorySource) Update(ctx context.Context) error {
	defer func() {
		n.IsLoading = false
		n.MustStore(ctx)
	}()
	n.IsLoading = true
	n.MustStore(ctx)

	workspaceIDs, err := LoadWorkspacesInSource(ctx, n.Directory, n.ID)
	if err != nil {
		return err
	}
	n.WorkspacesIDs = workspaceIDs
	n.MustStore(ctx)
	return nil
}

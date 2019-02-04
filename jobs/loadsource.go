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

package jobs

import (
	"context"

	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/relay"
)

// LoadSource loads the workspaces of the source and updates it.
func LoadSource(ctx context.Context, sourceID string, priority models.JobPriority) (string, error) {
	parts, err := relay.DecodeID(sourceID)
	if err != nil {
		return "", err
	}

	switch parts[0] {
	case models.NodeTypeDirectorySource:
		return LoadDirectorySource(ctx, sourceID, priority)
	case models.NodeTypeGitSource:
		return LoadGitSource(ctx, sourceID, priority)
	}

	return "", models.ErrType
}

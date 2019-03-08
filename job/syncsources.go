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

package job

import (
	"context"

	"groundcontrol/appcontext"
	"groundcontrol/model"
)

// SyncSources queues Jobs to sync all the Sources.
func SyncSources(ctx context.Context, highPriority bool) []string {
	appCtx := appcontext.Get(ctx)
	viewer := model.MustLoadUser(ctx, appCtx.ViewerID)
	var jobIDs []string
	for _, sourceID := range viewer.SourcesIDs {
		jobID, err := SyncSource(ctx, sourceID, highPriority)
		if err != nil {
			appCtx.Log.ErrorWithOwner(ctx, appCtx.SystemID, "SyncSources failed because %s", err.Error())
			continue
		}
		jobIDs = append(jobIDs, jobID)
	}
	return jobIDs
}

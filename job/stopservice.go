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

// StopService queues a Job to stop a Service.
func StopService(ctx context.Context, serviceID string, highPriority bool) (string, error) {
	if _, err := model.LoadService(ctx, serviceID); err != nil {
		return "", err
	}
	appCtx := appcontext.Get(ctx)
	service := model.MustLoadService(ctx, serviceID)
	return appCtx.Jobs.Add(ctx, JobNameStopService, service.WorkspaceID, highPriority, func(ctx context.Context) error {
		return appCtx.Services.Stop(ctx, serviceID)
	}), nil
}

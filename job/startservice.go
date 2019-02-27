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

	"groundcontrol/model"
)

// StartService starts a Service.
func StartService(ctx context.Context, serviceID string, env []string, priority model.JobPriority) (string, error) {
	if _, err := model.LoadService(ctx, serviceID); err != nil {
		return "", err
	}

	modelCtx := model.GetContext(ctx)

	return modelCtx.Jobs.Add(
		ctx,
		JobNameStartService,
		model.MustLoadService(ctx, serviceID).WorkspaceID,
		priority,
		func(ctx context.Context) error {
			return modelCtx.Services.Start(ctx, serviceID, env)
		},
	), nil
}

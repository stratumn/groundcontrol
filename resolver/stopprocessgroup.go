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

package resolver

import (
	"context"

	"groundcontrol/model"
)

func (r *mutationResolver) StopProcessGroup(ctx context.Context, id string) (*model.ProcessGroup, error) {
	modelCtx := model.GetModelContext(ctx)
	pm := modelCtx.PM

	processGroup, err := model.LoadProcessGroup(ctx, id)
	if err != nil {
		return nil, err
	}

	for _, processID := range processGroup.ProcessesIDs {
		process := model.MustLoadProcess(ctx, processID)

		if process.Status != model.ProcessStatusRunning {
			continue
		}

		if err := pm.Stop(ctx, process.ID); err != nil {
			return nil, err
		}
	}

	return processGroup, nil
}
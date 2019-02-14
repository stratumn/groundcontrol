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

package resolvers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"groundcontrol/models"
)

func (r *mutationResolver) OpenEditor(
	ctx context.Context,
	filename string,
) (models.Ok, error) {
	modelCtx := models.GetModelContext(ctx)

	command := fmt.Sprintf(modelCtx.OpenEditorCommand, filename)
	parts := strings.Split(command, " ")
	cmd := exec.Command(parts[0], parts[1:]...)

	if err := cmd.Run(); err != nil {
		return models.Ok{}, err
	}

	return models.Ok{Ok: true}, nil
}

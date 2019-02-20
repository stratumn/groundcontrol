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

	"groundcontrol/models"
)

type queryResolver struct {
	*Resolver
}

func (r *queryResolver) Node(ctx context.Context, id string) (models.Node, error) {
	node, _ := models.GetModelContext(ctx).Nodes.Load(id)
	return node, nil
}

func (r *queryResolver) Viewer(ctx context.Context) (models.User, error) {
	modelCtx := models.GetModelContext(ctx)
	return models.MustLoadUser(ctx, modelCtx.ViewerID), nil
}

func (r *queryResolver) System(ctx context.Context) (models.System, error) {
	modelCtx := models.GetModelContext(ctx)
	return models.MustLoadSystem(ctx, modelCtx.SystemID), nil
}

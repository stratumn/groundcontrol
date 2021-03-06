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

	"groundcontrol/appcontext"
	"groundcontrol/model"
)

func (r *mutationResolver) SetKey(ctx context.Context, input model.KeyInput) (*model.Key, error) {
	node := model.NewKey(input.Name, input.Value)
	if err := node.Store(ctx); err != nil {
		return nil, err
	}
	return node, appcontext.Get(ctx).Keys.Save()
}

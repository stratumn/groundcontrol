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

package groundcontrol

import (
	"context"
)

type queryResolver struct{ *Resolver }

func (r *queryResolver) Node(ctx context.Context, id string) (Node, error) {
	panic("not implemented")
}

func (r *queryResolver) Viewer(ctx context.Context) (User, error) {
	return r.Resolver.Viewer, nil
}

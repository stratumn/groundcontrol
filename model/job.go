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

import (
	"context"
	"fmt"
	"groundcontrol/appcontext"
)

// String is a string representation for the type instance.
func (n *Job) String() string {
	return n.Name
}

// LongString is a long string representation for the type instance.
func (n *Job) LongString(ctx context.Context) string {
	nodes := appcontext.Get(ctx).Nodes
	if n.OwnerID != "" {
		if owner, ok := nodes.Load(n.OwnerID); ok {
			if owner, ok := owner.(LongStringer); ok {
				return fmt.Sprintf("%s (%s)", n, owner.LongString(ctx))
			}
			return fmt.Sprintf("%s (%s)", n, owner)
		}
		return fmt.Sprintf("%s (%s)", n, n.OwnerID)
	}
	return fmt.Sprint(n)
}

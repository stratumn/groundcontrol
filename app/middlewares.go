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

package app

import (
	"context"

	"github.com/99designs/gqlgen/graphql"

	"groundcontrol/appcontext"
)

// modelContextMiddleware adds an appcontext.Context to the context that is
// passed to GraphQL resolvers. Currently gqlgen doesn't pass this context to
// subscriptions, so as a workaround the app context is also stored in the
// root resolver.
func modelContextMiddleware(appCtx *appcontext.Context) graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
		return next(appcontext.With(ctx, appCtx))
	}
}

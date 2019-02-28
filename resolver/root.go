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
	"groundcontrol/gql"
	"groundcontrol/model"
	"groundcontrol/store"
)

// Resolver is the root GraphQL resolver.
type Resolver struct {
	// We need this here because gqlgen doesn't currently pass it in the context of subscriptions.
	AppCtx *appcontext.Context
}

// Query returns the resolver for queries.
func (r *Resolver) Query() gql.QueryResolver {
	return &queryResolver{r}
}

// Mutation returns the resolver for mutations.
func (r *Resolver) Mutation() gql.MutationResolver {
	return &mutationResolver{r}
}

// Subscription returns the resolver for subscriptions.
func (r *Resolver) Subscription() gql.SubscriptionResolver {
	return &subscriptionResolver{r}
}

type mutationResolver struct {
	*Resolver
}

type subscriptionResolver struct {
	*Resolver
}

type queryResolver struct {
	*Resolver
}

func (r *queryResolver) Node(ctx context.Context, id string) (store.Node, error) {
	node, _ := appcontext.Get(ctx).Nodes.Load(id)
	return node, nil
}

func (r *queryResolver) Viewer(ctx context.Context) (*model.User, error) {
	appCtx := appcontext.Get(ctx)
	return model.LoadUser(ctx, appCtx.ViewerID)
}

func (r *queryResolver) System(ctx context.Context) (*model.System, error) {
	appCtx := appcontext.Get(ctx)
	return model.LoadSystem(ctx, appCtx.SystemID)
}

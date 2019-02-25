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
	"groundcontrol/gql"
	"groundcontrol/model"
)

// Resolver is the root GraphQL resolver.
type Resolver struct {
	// We need this here because gqlgen doesn't currently pass it in the context of subscriptions.
	ModelCtx *model.Context
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

func (r *queryResolver) Node(ctx context.Context, id string) (model.Node, error) {
	node, _ := model.GetContext(ctx).Nodes.Load(id)
	return node, nil
}

func (r *queryResolver) Viewer(ctx context.Context) (*model.User, error) {
	modelCtx := model.GetContext(ctx)
	return model.LoadUser(ctx, modelCtx.ViewerID)
}

func (r *queryResolver) System(ctx context.Context) (*model.System, error) {
	modelCtx := model.GetContext(ctx)
	return model.LoadSystem(ctx, modelCtx.SystemID)
}

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
	"github.com/stratumn/groundcontrol/gql"
	"github.com/stratumn/groundcontrol/jobs"
	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/pubsub"
)

// Resolver is the root GraphQL resolver.
type Resolver struct {
	Nodes               *models.NodeManager
	Log                 *models.Logger
	Jobs                *models.JobManager
	PM                  *models.ProcessManager
	Subs                *pubsub.PubSub
	GetProjectPath      models.ProjectPathGetter
	GetProjectCachePath jobs.ProjectCachePathGetter
	ViewerID            string
	SystemID            string
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

// User returns the resolver for a user.
func (r *Resolver) User() gql.UserResolver {
	return &userResolver{r}
}

// Workspace returns the resolver for a workspace.
func (r *Resolver) Workspace() gql.WorkspaceResolver {
	return &workspaceResolver{r}
}

// Project returns the resolver for a project.
func (r *Resolver) Project() gql.ProjectResolver {
	return &projectResolver{r}
}

// Task returns the resolver for a task.
func (r *Resolver) Task() gql.TaskResolver {
	return &taskResolver{r}
}

// Step returns the resolver for a step.
func (r *Resolver) Step() gql.StepResolver {
	return &stepResolver{r}
}

// System returns the resolver for system data.
func (r *Resolver) System() gql.SystemResolver {
	return &systemResolver{r}
}

// Job returns the resolver for a job.
func (r *Resolver) Job() gql.JobResolver {
	return &jobResolver{r}
}

// ProcessGroup returns the resolver for a process group.
func (r *Resolver) ProcessGroup() gql.ProcessGroupResolver {
	return &processGroupResolver{r}
}

// Process returns the resolver for a process.
func (r *Resolver) Process() gql.ProcessResolver {
	return &processResolver{r}
}

// LogEntry returns the resolver for a log entry.
func (r *Resolver) LogEntry() gql.LogEntryResolver {
	return &logEntryResolver{r}
}

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
	"path"
	"path/filepath"

	"github.com/stratumn/groundcontrol/gql"
	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/relay"
)

// Resolver is the root GraphQL resolver.
type Resolver struct {
	Nodes          *models.NodeManager
	Log            *models.Logger
	Jobs           *models.JobManager
	Subs           *pubsub.PubSub
	GetProjectPath models.ProjectPathGetter
	ViewerID       string
	SystemID       string
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

// Job returns the resolver for a job.
func (r *Resolver) Job() gql.JobResolver {
	return &jobResolver{r}
}

// System returns the resolver for system data.
func (r *Resolver) System() gql.SystemResolver {
	return &systemResolver{r}
}

// CreateResolver creates a resolver from a config file.
func CreateResolver(filename string) (*Resolver, error) {
	config, err := models.LoadConfigYAML(filename)
	if err != nil {
		return nil, err
	}

	nodes := &models.NodeManager{}
	viewer, err := config.CreateNodes(nodes)
	if err != nil {
		return nil, err
	}

	logMetricsID := relay.EncodeID(models.NodeTypeLogMetrics, filename)
	systemID := relay.EncodeID(models.NodeTypeSystem, filename)
	jobMetricsID := relay.EncodeID(models.NodeTypeJobMetrics, filename)

	nodes.MustStoreLogMetrics(models.LogMetrics{
		ID: logMetricsID,
	})

	nodes.MustStoreJobMetrics(models.JobMetrics{
		ID: jobMetricsID,
	})

	nodes.MustStoreSystem(models.System{
		ID:           systemID,
		JobMetricsID: jobMetricsID,
		LogMetricsID: logMetricsID,
	})

	subs := pubsub.New()
	log := models.NewLogger(nodes, subs, 100, models.LogLevelDebug, systemID)
	jobs := models.NewJobManager(nodes, log, subs, 2, systemID)

	return &Resolver{
		Nodes: nodes,
		Log:   log,
		Jobs:  jobs,
		Subs:  subs,
		GetProjectPath: func(workspaceSlug, repo, branch string) string {
			name := path.Base(repo)
			ext := path.Ext(name)
			name = name[:len(name)-len(ext)]
			return filepath.Join("workspaces", workspaceSlug, name)
		},
		ViewerID: viewer.ID,
		SystemID: systemID,
	}, nil
}

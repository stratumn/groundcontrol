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

	"github.com/stratumn/groundcontrol/models"
)

type subscriptionResolver struct{ *Resolver }

func (r *subscriptionResolver) WorkspaceUpdated(ctx context.Context, id *string) (<-chan models.Workspace, error) {
	ch := make(chan models.Workspace)

	unsubscribe := models.SubscribeWorkspaceUpdated(func(workspace *models.Workspace) {
		if id != nil && *id != workspace.ID {
			return
		}

		ch <- *workspace
	})

	go func() {
		<-ctx.Done()
		unsubscribe()
		for len(ch) > 0 {
			<-ch
		}
	}()

	return ch, nil
}
func (r *subscriptionResolver) ProjectUpdated(ctx context.Context, id *string) (<-chan models.Project, error) {
	ch := make(chan models.Project)

	unsubscribe := models.SubscribeProjectUpdated(func(project *models.Project) {
		if id != nil && *id != project.ID {
			return
		}

		ch <- *project
	})

	go func() {
		<-ctx.Done()
		unsubscribe()
		for len(ch) > 0 {
			<-ch
		}
	}()

	return ch, nil
}
func (r *subscriptionResolver) JobUpserted(ctx context.Context) (<-chan models.Job, error) {
	ch := make(chan models.Job)

	unsubscribe := r.JobManager.Subscribe(func(job *models.Job) {
		ch <- *job
	})

	go func() {
		<-ctx.Done()
		unsubscribe()
		for len(ch) > 0 {
			<-ch
		}
	}()

	return ch, nil
}

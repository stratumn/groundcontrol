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

type subscriptionResolver struct{ *Resolver }

func (r *subscriptionResolver) WorkspaceUpdated(ctx context.Context, id *string) (<-chan Workspace, error) {
	ch := make(chan Workspace)

	unsubscribe := SubscribeWorkspaceUpdated(func(workspace *Workspace) {
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
func (r *subscriptionResolver) ProjectUpdated(ctx context.Context, id *string) (<-chan Project, error) {
	ch := make(chan Project)

	unsubscribe := SubscribeProjectUpdated(func(project *Project) {
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
func (r *subscriptionResolver) JobUpserted(ctx context.Context) (<-chan Job, error) {
	ch := make(chan Job)

	unsubscribe := SubscribeJobUpserted(func(job *Job) {
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

//go:generate go run ./scripts/gqlgen.go
package groundcontrol

import (
	"context"
)

type Resolver struct{}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}
func (r *Resolver) Subscription() SubscriptionResolver {
	return &subscriptionResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CloneProject(ctx context.Context, id string) (Job, error) {
	panic("not implemented")
}
func (r *mutationResolver) CloneWorkspace(ctx context.Context, id string) ([]Job, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Node(ctx context.Context, id string) (Node, error) {
	panic("not implemented")
}
func (r *queryResolver) Viewer(ctx context.Context) (User, error) {
	return Viewer, nil
}

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

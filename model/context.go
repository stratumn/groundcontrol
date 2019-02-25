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

	"groundcontrol/pubsub"
)

type contextKey string

const ContextKey contextKey = "model_context"

// ProjectGitSourcePathGetter is a function that returns the path to a Git source.
type ProjectGitSourcePathGetter func(repo, reference string) string

// ProjectPathGetter is a function that returns the path to a project.
type ProjectPathGetter func(workspaceSlug, projectSlug string) string

// ProjectCachePathGetter is a function that returns the path to a project's cache.
type ProjectCachePathGetter func(workspaceSlug, projectSlug string) string

// Context contains variables that are passed to model functions.
type Context struct {
	Nodes               *NodeManager
	Log                 *Logger
	Jobs                *JobManager
	PM                  *ProcessManager
	Subs                *pubsub.PubSub
	Sources             *SourcesConfig
	Keys                *KeysConfig
	GetGitSourcePath    ProjectGitSourcePathGetter
	GetProjectPath      ProjectPathGetter
	GetProjectCachePath ProjectCachePathGetter
	OpenEditorCommand   string
	ViewerID            string
	SystemID            string
}

// WithContext adds a model context to a Go context.
func WithContext(ctx context.Context, mc *Context) context.Context {
	return context.WithValue(ctx, ContextKey, mc)
}

// GetContext retrieves the model context from a Go context.
func GetContext(ctx context.Context) *Context {
	if val, ok := ctx.Value(ContextKey).(*Context); ok {
		return val
	}

	return nil
}

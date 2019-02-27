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
)

type contextKey string

const modelCtxKey contextKey = "model_context"

// Context contains variables that are passed to model functions.
type Context struct {
	Nodes               Nodes
	Log                 Log
	Jobs                Jobs
	Services            Services
	Subs                Subs
	Sources             Sources
	Keys                Keys
	GetGitSourcePath    ProjectGitSourcePathGetter
	GetProjectPath      ProjectPathGetter
	GetProjectCachePath ProjectCachePathGetter
	OpenEditorCommand   string
	ViewerID            string
	SystemID            string
}

// WithContext adds a model context to a Go context.
func WithContext(ctx context.Context, mc *Context) context.Context {
	return context.WithValue(ctx, modelCtxKey, mc)
}

// GetContext retrieves the model context from a Go context.
func GetContext(ctx context.Context) *Context {
	if val, ok := ctx.Value(modelCtxKey).(*Context); ok {
		return val
	}

	return nil
}

// Nodes exposes low-level functions to load, store, and lock nodes.
type Nodes interface {
	// Store stores a node.
	Store(id string, node Node)
	// Load loads a node.
	Load(id string) (Node, bool)
	// MustLoad loads a node or panics if it doesn't exist.
	MustLoad(id string) Node
	// Delete deletes a node.
	Delete(id string)
	// Lock locks the given IDs.
	Lock(ids ...string)
	// Unlock unlocks the given IDs.
	Unlock(ids ...string)
}

// Log exposes functions to emit log messages.
type Log interface {
	// Add adds a log entry.
	Add(ctx context.Context, level LogLevel, ownerID string, message string) (string, error)
	// Debug adds a debug entry.
	Debug(ctx context.Context, message string, a ...interface{}) string
	// Info adds an info entry.
	Info(ctx context.Context, message string, a ...interface{}) string
	// Warning adds a warning entry.
	Warning(ctx context.Context, message string, a ...interface{}) string
	// Error adds an error entry.
	Error(ctx context.Context, message string, a ...interface{}) string
	// DebugWithOwner adds a debug entry with an owner.
	DebugWithOwner(ctx context.Context, ownerID string, message string, a ...interface{}) string
	// InfoWithOwner adds an info entry with an owner.
	InfoWithOwner(ctx context.Context, ownerID string, message string, a ...interface{}) string
	// WarningWithOwner adds a warning entry with an owner.
	WarningWithOwner(ctx context.Context, ownerID string, message string, a ...interface{}) string
	// ErrorWithOwner adds an error entry with an owner.
	ErrorWithOwner(ctx context.Context, ownerID string, message string, a ...interface{}) string
}

// Jobs exposes functions to queue jobs.
type Jobs interface {
	Work(ctx context.Context) error
	// Add adds a job to the queue and returns the job's ID.
	Add(ctx context.Context, name string, ownerID string, priority JobPriority, fn func(ctx context.Context) error) string
	// Stop cancels a running job.
	Stop(ctx context.Context, id string) error
}

// Services exposes functions to start and stop services.
type Services interface {
	// Start starts a Service and its dependencies.
	Start(ctx context.Context, serviceID string, env []string) error
	// Stop stops a running Service.
	Stop(ctx context.Context, serviceID string) error
	// Clean terminates all running Services.
	Clean(ctx context.Context)
}

// Subs exposes functions to subscribe and publish messages.
type Subs interface {
	Subscribe(ctx context.Context, messageType string, since uint64, fn func(interface{}))
	// Publish will publish a message of the given type to all subscribers for that type.
	Publish(messageType string, message interface{})
	// LastMessageID returns the ID of the last message.
	LastMessageID() uint64
}

// Sources exposes functions to load and store sources to disk.
type Sources interface {
	// Store stores nodes for the content of the sources config.
	Store(ctx context.Context) error
	// SetDirectorySource sets a directory source and stores the corresponding node.
	// It returns the ID of the source.
	SetDirectorySource(ctx context.Context, input DirectorySourceInput) string
	// SetGitSource sets a Git source and stores the corresponding node.
	// It returns the ID of the source.
	SetGitSource(ctx context.Context, input GitSourceInput) string
	// Delete deletes a source.
	Delete(ctx context.Context, id string) error
	// Save saves the config to disk, overwriting the file if it exists.
	Save() error
}

// Keys exposes functions to load and store keys to disk.
type Keys interface {
	// Store stores nodes for the content of the keys config.
	Store(ctx context.Context) error
	// Set sets a key and stores the corresponding node.
	// It returns the ID of the key.
	Set(ctx context.Context, input KeyInput) string
	// Delete deletes a key and the corresponding node.
	Delete(ctx context.Context, id string) error
	// Save saves the config to disk, overwriting the file if it exists.
	Save() error
}

// ProjectGitSourcePathGetter is a function that returns the path to a Git source.
type ProjectGitSourcePathGetter func(repo, reference string) string

// ProjectPathGetter is a function that returns the path to a project.
type ProjectPathGetter func(workspaceSlug, projectSlug string) string

// ProjectCachePathGetter is a function that returns the path to a project's cache.
type ProjectCachePathGetter func(workspaceSlug, projectSlug string) string

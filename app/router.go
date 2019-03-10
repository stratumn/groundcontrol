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
	"net/http"

	"github.com/99designs/gqlgen-contrib/gqlapollotracing"
	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"

	"groundcontrol/appcontext"
	"groundcontrol/gql"
	"groundcontrol/resolver"
)

// router offloads setting up HTTP middlewares and routes for the App.
type router struct {
	*chi.Mux
}

// newRouter creates a new router.
func newRouter() router {
	return router{Mux: chi.NewRouter()}
}

// EnableCORS adds CORS headers to enable cross-origin requests from any
// domain. It must be called before routes are added.
func (r router) EnableCORS() {
	opts := cors.Options{AllowCredentials: true, Debug: false}
	handler := cors.New(opts).Handler
	r.Use(handler)
}

// EnableGQL adds routes for GraphQL. Apollo tracing can be enabled, but it
// currently crashes with subscriptions.
func (r *router) EnableGQL(appCtx *appcontext.Context, enableApolloTracing bool) {
	root := resolver.Resolver{AppCtx: appCtx}
	config := gql.Config{Resolvers: &root}
	opts := []handler.Option{
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		}),
		handler.ResolverMiddleware(modelContextMiddleware(appCtx)),
	}
	if enableApolloTracing {
		opts = append(
			opts,
			handler.RequestMiddleware(gqlapollotracing.RequestMiddleware()),
			handler.Tracer(gqlapollotracing.NewTracer()),
		)
	}
	schema := gql.NewExecutableSchema(config)
	graphql := handler.GraphQL(schema, opts...)
	r.Handle("/query", graphql)
}

// EnablePlayground adds a route for the GraphQL playground user interface.
func (r router) EnablePlayground() {
	r.Handle("/graphql", handler.Playground("GraphQL playground", "/query"))
}

// EnableUI, given a file system containing the static assets, adds a route for
// the user interface.
func (r *router) EnableUI(ui http.FileSystem) {
	fileServer := http.FileServer(ui)
	// Handle all requests that don't have a route.
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		if _, err := ui.Open(req.URL.Path); err != nil {
			// If a file doesn't exist, rewrite the path for pushstate.
			req.URL.Path = ""
		}
		fileServer.ServeHTTP(w, req)
	})
}

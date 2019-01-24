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

//go:generate go run scripts/nodesgen.go -t User,Workspace,Project,Commit,System,Job,JobMetrics,LogEntry,LogMetrics,Task,Step -o models/auto_nodes.go
//go:generate go run scripts/paginatorsgen.go -t Commit,Job,LogEntry -o models/auto_paginators.go
//go:generate go run scripts/gqlgen.go

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/stratumn/groundcontrol/gql"
	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/resolvers"
)

const defaultPort = "8080"

var ui http.FileSystem

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func logMiddleware(log *models.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug("Request", struct {
				Method     string
				URL        string
				RemoteAddr string
			}{
				r.Method,
				r.URL.String(),
				r.RemoteAddr,
			})

			h.ServeHTTP(w, r)
		})
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	args := os.Args
	if len(args) > 2 {
		fmt.Printf("usage: %s [file]\n", args[0])
		os.Exit(1)
	}

	filename, err := filepath.Abs("groundcontrol.yml")
	checkError(err)

	if len(args) > 1 {
		filename, err = filepath.Abs(args[1])
		checkError(err)
	}

	resolver, err := resolvers.CreateResolver(filename)
	checkError(err)

	go resolver.Jobs.Work(context.Background())

	gqlConfig := gql.Config{
		Resolvers: resolver,
	}

	c := cors.New(cors.Options{
		AllowCredentials: true,
		Debug:            false,
	})

	router := chi.NewRouter()
	router.Use(logMiddleware(resolver.Log), c.Handler)

	if ui != nil {
		fileServer := http.FileServer(ui)
		router.NotFound(func(w http.ResponseWriter, req *http.Request) {
			if _, err := ui.Open(req.URL.Path); err != nil {
				// Rewrite other URLs to index for pushstate.
				req.URL.Path = ""
			}
			fileServer.ServeHTTP(w, req)
		})
	} else {
		router.Handle("/", handler.Playground("GraphQL playground", "/query"))
	}

	router.Handle("/query", handler.GraphQL(
		gql.NewExecutableSchema(gqlConfig),
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		}),
	))

	if ui != nil {
		resolver.Log.Info("App Ready", struct {
			UserInterfaceURL string
		}{
			fmt.Sprintf("http://localhost:%s", port),
		})
	} else {
		resolver.Log.Info("App Ready", struct {
			GraphQLPlaygroundURL string
		}{
			fmt.Sprintf("http://localhost:%s", port),
		})
	}

	if err := http.ListenAndServe(":"+port, router); err != nil {
		resolver.Log.Error("App Crashed", struct {
			Error error
		}{
			err,
		})
	}
}

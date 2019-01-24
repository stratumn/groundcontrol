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

//go:generate go run scripts/nodesgen.go -t User,Workspace,Project,Commit,System,Job,JobMetrics,LogEntry,LogMetrics -o models/auto_nodes.go
//go:generate go run scripts/paginatorsgen.go -t Commit,Job,LogEntry -o models/auto_paginators.go
//go:generate go run scripts/gqlgen.go

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"github.com/stratumn/groundcontrol/gql"
	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/relay"
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

	config, err := models.LoadConfigYAML(filename)
	checkError(err)

	nodes := &models.NodeManager{}
	viewer, err := config.CreateNodes(nodes)
	checkError(err)

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

	resolver := resolvers.Resolver{
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
	}

	go jobs.Work(context.Background())

	gqlConfig := gql.Config{
		Resolvers: &resolver,
	}

	c := cors.New(cors.Options{
		AllowCredentials: true,
		Debug:            false,
	})

	router := chi.NewRouter()
	router.Use(c.Handler)

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
		log.Info("App Ready", struct {
			UserInterfaceURL string
		}{
			fmt.Sprintf("http://localhost:%s", port),
		})
	} else {
		log.Info("App Ready", struct {
			GraphQLPlaygroundURL string
		}{
			fmt.Sprintf("http://localhost:%s", port),
		})
	}

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Error("App Crashed", struct {
			Error error
		}{
			err,
		})
	}
}

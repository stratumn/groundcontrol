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
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/pkg/browser"
	"github.com/rs/cors"

	"github.com/stratumn/groundcontrol/gql"
	"github.com/stratumn/groundcontrol/jobs"
	"github.com/stratumn/groundcontrol/models"
	"github.com/stratumn/groundcontrol/pubsub"
	"github.com/stratumn/groundcontrol/relay"
	"github.com/stratumn/groundcontrol/resolvers"
)

// App contains data about the app.
type App struct {
	configFilenames         []string
	listenAddress           string
	jobConcurrency          int
	logLevel                models.LogLevel
	logCap                  int
	checkProjectsInterval   time.Duration
	disableSignalHandling   bool
	gracefulShutdownTimeout time.Duration
	ui                      http.FileSystem
	openBrowser             bool
	projectPathGetter       models.ProjectPathGetter
	projectCachePathGetter  jobs.ProjectCachePathGetter
}

// New creates a new App.
func New(opts ...Opt) *App {
	app := &App{
		listenAddress:           DefaultListenAddress,
		jobConcurrency:          DefaultJobConcurrency,
		logLevel:                DefaultLogLevel,
		logCap:                  DefaultLogCap,
		checkProjectsInterval:   DefaultCheckProjectsInterval,
		gracefulShutdownTimeout: DefaultGracefulShutdownTimeout,
		openBrowser:             DefaultOpenBrowser,
		projectPathGetter:       DefaultProjectPathGetter,
		projectCachePathGetter:  DefaultProjectCachePathGetter,
	}

	for _, opt := range opts {
		opt(app)
	}

	if len(app.configFilenames) < 1 {
		app.configFilenames = []string{DefaultConfigFilename}
	}

	return app
}

// Start starts the app. It blocks until an error occurs or the app exits.
func (a *App) Start(ctx context.Context) error {
	nodes := &models.NodeManager{}

	viewerID, systemID := a.createBaseNodes(nodes)

	subs := pubsub.New()
	log := models.NewLogger(nodes, subs, a.logCap, a.logLevel, systemID)
	jobs := models.NewJobManager(nodes, log, subs, a.jobConcurrency, systemID)
	pm := models.NewProcessManager(nodes, log, subs, a.projectPathGetter, systemID)

	if err := a.loadConfigs(nodes, viewerID); err != nil {
		return err
	}

	router := chi.NewRouter()

	if a.logLevel <= models.LogLevelDebug {
		router.Use(logMiddleware(log))
	}

	corsHandler := cors.New(cors.Options{
		AllowCredentials: true,
		Debug:            false,
	}).Handler

	router.Use(corsHandler)

	router.Handle("/graphql", handler.Playground("GraphQL playground", "/query"))

	if a.ui != nil {
		fileServer := http.FileServer(a.ui)
		router.NotFound(func(w http.ResponseWriter, req *http.Request) {
			if _, err := a.ui.Open(req.URL.Path); err != nil {
				// Rewrite other URLs to index for pushstate.
				req.URL.Path = ""
			}
			fileServer.ServeHTTP(w, req)
		})
	}

	resolver := &resolvers.Resolver{
		Nodes:               nodes,
		Log:                 log,
		Jobs:                jobs,
		PM:                  pm,
		Subs:                subs,
		GetProjectPath:      a.projectPathGetter,
		GetProjectCachePath: a.projectCachePathGetter,
		ViewerID:            viewerID,
		SystemID:            systemID,
	}

	gqlConfig := gql.Config{
		Resolvers: resolver,
	}

	router.Handle("/query", handler.GraphQL(
		gql.NewExecutableSchema(gqlConfig),
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		}),
	))

	server := &http.Server{
		Addr:    a.listenAddress,
		Handler: router,
	}

	go resolver.Jobs.Work(ctx)
	go a.startPeriodicJobs(ctx, nodes, log, jobs, subs, viewerID)
	if !a.disableSignalHandling {
		go a.handleSignals(ctx, log, pm, server)
	}
	go func() {
		log.Info("app ready")
		if a.ui != nil {
			log.Info("user interface on %s", a.listenAddress)
		}
		log.Info("GraphQL playground on %s/graphql", a.listenAddress)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("app crashed because %s", err.Error())
		}
	}()

	if a.openBrowser && a.ui != nil {
		a.openAddressInBrowser(log)
	}

	<-ctx.Done()

	return ctx.Err()
}

func (a *App) createBaseNodes(nodes *models.NodeManager) (string, string) {
	unique := strings.Join(a.configFilenames, ";")

	var (
		viewerID         = relay.EncodeID(models.NodeTypeUser, unique)
		systemID         = relay.EncodeID(models.NodeTypeSystem, unique)
		jobMetricsID     = relay.EncodeID(models.NodeTypeJobMetrics, unique)
		processMetricsID = relay.EncodeID(models.NodeTypeProcessMetrics, unique)
		logMetricsID     = relay.EncodeID(models.NodeTypeLogMetrics, unique)
	)

	nodes.MustStoreUser(models.User{
		ID: relay.EncodeID(models.NodeTypeUser, unique),
	})

	nodes.MustStoreLogMetrics(models.LogMetrics{
		ID: logMetricsID,
	})

	nodes.MustStoreJobMetrics(models.JobMetrics{
		ID: jobMetricsID,
	})

	nodes.MustStoreProcessMetrics(models.ProcessMetrics{
		ID: processMetricsID,
	})

	nodes.MustStoreSystem(models.System{
		ID:               systemID,
		JobMetricsID:     jobMetricsID,
		LogMetricsID:     logMetricsID,
		ProcessMetricsID: processMetricsID,
	})

	return viewerID, systemID
}

func (a *App) loadConfigs(nodes *models.NodeManager, viewerID string) error {
	for _, filename := range a.configFilenames {
		config, err := models.LoadConfigYAML(filename)
		if err != nil {
			return err
		}

		err = config.CreateNodes(nodes, viewerID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) startPeriodicJobs(
	ctx context.Context,
	nodes *models.NodeManager,
	log *models.Logger,
	jobManager *models.JobManager,
	subs *pubsub.PubSub,
	viewerID string,
) {
	jobs.StartPeriodic(
		ctx,
		nodes,
		subs,
		a.checkProjectsInterval,
		func() []string {
			return jobs.LoadAllCommits(
				nodes,
				log,
				jobManager,
				subs,
				a.projectPathGetter,
				a.projectCachePathGetter,
				viewerID,
			)
		},
	)
}

func (a *App) handleSignals(
	ctx context.Context,
	log *models.Logger,
	pm *models.ProcessManager,
	server *http.Server,
) {
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGTERM)
	signal.Notify(signalCh, syscall.SIGINT)

	log.Debug("received signal %d", <-signalCh)
	log.Info("starting graceful shutdown")

	a.shutdownGracefully(ctx, log, pm, server)
}

func (a *App) shutdownGracefully(
	ctx context.Context,
	log *models.Logger,
	pm *models.ProcessManager,
	server *http.Server,
) {
	shutdownCtx, cancel := context.WithTimeout(ctx, a.gracefulShutdownTimeout)
	defer cancel()

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(2)

	go func() {
		pm.Clean(shutdownCtx)
		waitGroup.Done()
	}()

	go func() {
		server.Shutdown(shutdownCtx)
		waitGroup.Done()
	}()

	waitGroup.Wait()

	if err := shutdownCtx.Err(); err != nil {
		log.Error("graceful shutdown failed because %s", err.Error())
		os.Exit(1)
	}

	log.Info("graceful shutdown complete, goodbye!")
	os.Exit(0)
}

func logMiddleware(log *models.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Debug("%s %s %s", r.Method, r.URL.String(), r.RemoteAddr)
			h.ServeHTTP(w, r)
		})
	}
}

func (a *App) openAddressInBrowser(log *models.Logger) {
	addr, err := net.ResolveTCPAddr("tcp", a.listenAddress)
	if err != nil {
		log.Warning("could not resolve address because %s", err.Error())
	} else {
		url := "http://"
		if addr.IP == nil {
			url += "localhost"
		} else {
			url += addr.IP.String()
		}
		if addr.Port != 0 {
			url += fmt.Sprintf(":%d", addr.Port)
		}
		if err := browser.OpenURL(url); err != nil {
			log.Warning("could not resolve address because %s", err.Error())
		}
	}
}

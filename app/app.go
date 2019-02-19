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
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/99designs/gqlgen-contrib/gqlapollotracing"
	"github.com/99designs/gqlgen/handler"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/pkg/browser"
	"github.com/rs/cors"

	"groundcontrol/gql"
	"groundcontrol/jobs"
	"groundcontrol/models"
	"groundcontrol/pubsub"
	"groundcontrol/relay"
	"groundcontrol/resolvers"
)

// App starts Ground Control.
type App struct {
	sourcesFile             string
	keysFile                string
	listenAddress           string
	jobConcurrency          int
	logLevel                models.LogLevel
	logCap                  int
	pubSubHistoryCap        int
	periodicJobsInterval    time.Duration
	gracefulShutdownTimeout time.Duration
	ui                      http.FileSystem
	openBrowser             bool
	gitSourcesDirectory     string
	workspacesDirectory     string
	cacheDirectory          string
	openEditorCommand       string
	enableApolloTracing     bool
	enableSignalHandling    bool
}

// New creates a new App with given options.
func New(opts ...Opt) *App {
	app := &App{
		sourcesFile:             DefaultSourcesFile,
		keysFile:                DefaultKeysFile,
		listenAddress:           DefaultListenAddress,
		jobConcurrency:          DefaultJobConcurrency,
		logLevel:                DefaultLogLevel,
		logCap:                  DefaultLogCap,
		pubSubHistoryCap:        DefaultPubSubHistoryCap,
		periodicJobsInterval:    DefaultPeriodicJobsInterval,
		gracefulShutdownTimeout: DefaultGracefulShutdownTimeout,
		openBrowser:             DefaultOpenBrowser,
		gitSourcesDirectory:     DefaultGitSourcesDirectory,
		workspacesDirectory:     DefaultWorkspacesDirectory,
		cacheDirectory:          DefaultCacheDirectory,
		openEditorCommand:       DefaultOpenEditorCommand,
		enableApolloTracing:     DefaultEnableApolloTracing,
		enableSignalHandling:    DefaultEnableSignalHandling,
	}

	for _, opt := range opts {
		opt(app)
	}

	return app
}

// Start starts the app. It blocks until an error occurs or the app exits.
func (a *App) Start(ctx context.Context) error {
	nodes := &models.NodeManager{}
	viewerID, systemID := a.createBaseNodes(nodes)
	subs := pubsub.New(a.pubSubHistoryCap)
	log := models.NewLogger(
		nodes,
		subs,
		a.logCap,
		a.logLevel,
		systemID,
		a.getProjectPath,
	)
	jobs := models.NewJobManager(a.jobConcurrency)
	pm := models.NewProcessManager()

	log.InfoWithOwner(systemID, "starting app")
	log.InfoWithOwner(
		systemID,
		"runtime %s %s %s",
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
	)

	sources, err := a.loadSources(nodes, subs, viewerID)
	if err != nil {
		return err
	}

	keys, err := a.loadKeys(nodes, subs, viewerID)
	if err != nil {
		return err
	}

	modelCtx := &models.ModelContext{
		Nodes:               nodes,
		Log:                 log,
		Jobs:                jobs,
		PM:                  pm,
		Subs:                subs,
		Sources:             sources,
		Keys:                keys,
		GetGitSourcePath:    a.getGitSourcePath,
		GetProjectPath:      a.getProjectPath,
		GetProjectCachePath: a.getProjectCachePath,
		OpenEditorCommand:   a.openEditorCommand,
		ViewerID:            viewerID,
		SystemID:            systemID,
	}

	ctx = models.WithModelContext(ctx, modelCtx)

	if err := initHooks(ctx); err != nil {
		return err
	}

	router := chi.NewRouter()

	if a.logLevel <= models.LogLevelDebug {
		router.Use(logHTTPRequestMiddleware(log, systemID))
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

	gqlConfig := gql.Config{
		Resolvers: &resolvers.Resolver{
			Nodes: nodes,
			Subs:  subs,
		},
	}

	gqlOptions := []handler.Option{
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		}),
		handler.ResolverMiddleware(modelContextResolverMiddleware(modelCtx)),
	}

	if a.enableApolloTracing {
		gqlOptions = append(
			gqlOptions,
			handler.RequestMiddleware(gqlapollotracing.RequestMiddleware()),
			handler.Tracer(gqlapollotracing.NewTracer()),
		)
	}

	router.Handle("/query", handler.GraphQL(
		gql.NewExecutableSchema(gqlConfig),
		gqlOptions...,
	))

	server := &http.Server{
		Addr:    a.listenAddress,
		Handler: router,
	}

	go func() {
		if err := jobs.Work(ctx); err != nil && err != context.Canceled {
			log.ErrorWithOwner(
				systemID,
				"job manager crashed because %s",
				err.Error(),
			)
		}
	}()

	a.startPeriodicJobs(ctx, log, systemID)
	if a.enableSignalHandling {
		go a.handleSignals(ctx, log, pm, server, systemID)
	}

	errorCh := make(chan error, 1)

	go func() {
		log.InfoWithOwner(systemID, "app ready")
		if a.ui != nil {
			log.InfoWithOwner(systemID, "user interface on %s", a.listenAddress)
		}
		log.InfoWithOwner(
			systemID,
			"GraphQL playground on %s/graphql",
			a.listenAddress,
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errorCh <- err
		}
	}()

	if a.openBrowser && a.ui != nil {
		a.openAddressInBrowser(log, systemID)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errorCh:
		return err
	}
}

func (a *App) createBaseNodes(nodes *models.NodeManager) (string, string) {
	var (
		viewerID         = relay.EncodeID(models.NodeTypeUser)
		systemID         = relay.EncodeID(models.NodeTypeSystem)
		jobMetricsID     = relay.EncodeID(models.NodeTypeJobMetrics)
		processMetricsID = relay.EncodeID(models.NodeTypeProcessMetrics)
		logMetricsID     = relay.EncodeID(models.NodeTypeLogMetrics)
	)

	nodes.MustStoreUser(models.User{
		ID: viewerID,
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

func (a *App) loadSources(
	nodes *models.NodeManager,
	subs *pubsub.PubSub,
	viewerID string,
) (*models.SourcesConfig, error) {
	config, err := models.LoadSourcesConfigYAML(a.sourcesFile)
	if err != nil {
		return nil, err
	}

	err = config.UpsertNodes(nodes, subs, viewerID)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (a *App) loadKeys(
	nodes *models.NodeManager,
	subs *pubsub.PubSub,
	viewerID string,
) (*models.KeysConfig, error) {
	config, err := models.LoadKeysConfigYAML(a.keysFile)
	if err != nil {
		return nil, err
	}

	err = config.UpsertNodes(nodes, subs, viewerID)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (a *App) startPeriodicJobs(ctx context.Context, log *models.Logger, systemID string) {
	go func() {
		err := jobs.StartPeriodic(
			ctx,
			a.periodicJobsInterval,
			jobs.LoadAllSources,
			jobs.LoadAllCommits,
		)
		if err != nil && err != context.Canceled {
			log.ErrorWithOwner(
				systemID,
				"job manager crashed because %s",
				err.Error(),
			)
		}
	}()

}

func (a *App) handleSignals(
	ctx context.Context,
	log *models.Logger,
	pm *models.ProcessManager,
	server *http.Server,
	systemID string,
) {
	signalCh := make(chan os.Signal, 2)
	signal.Notify(signalCh, syscall.SIGTERM)
	signal.Notify(signalCh, syscall.SIGINT)

	log.DebugWithOwner(systemID, "received signal %d", <-signalCh)
	log.InfoWithOwner(systemID, "starting graceful shutdown")

	a.shutdownGracefully(ctx, log, pm, server, systemID)
}

func (a *App) shutdownGracefully(
	ctx context.Context,
	log *models.Logger,
	pm *models.ProcessManager,
	server *http.Server,
	systemID string,
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
		if err := server.Shutdown(ctx); err != nil && err != context.Canceled {
			log.ErrorWithOwner(
				systemID,
				"server shutdown failed because %s",
				err.Error(),
			)
		}
		waitGroup.Done()
	}()

	waitGroup.Wait()

	if err := shutdownCtx.Err(); err != nil {
		log.ErrorWithOwner(
			systemID,
			"graceful shutdown failed because %s",
			err.Error(),
		)
		os.Exit(1)
	}

	log.InfoWithOwner(systemID, "graceful shutdown complete, goodbye!")
	os.Exit(0)
}

func (a *App) openAddressInBrowser(log *models.Logger, systemID string) {
	addr, err := net.ResolveTCPAddr("tcp", a.listenAddress)
	if err != nil {
		log.WarningWithOwner(
			systemID,
			"could not resolve address because %s",
			err.Error(),
		)
		return
	}

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
		log.WarningWithOwner(
			systemID,
			"could not resolve address because %s",
			err.Error(),
		)
	}
}

func (a *App) getGitSourcePath(repo, reference string) string {
	name := path.Base(repo)
	ext := path.Ext(name)
	name = name[:len(name)-len(ext)]
	return filepath.Join(a.gitSourcesDirectory, name, reference)
}

func (a *App) getProjectPath(workspaceSlug, projectSlug string) string {
	return filepath.Join(a.workspacesDirectory, workspaceSlug, projectSlug)
}

func (a *App) getProjectCachePath(workspaceSlug, projectSlug string) string {
	return filepath.Join(a.cacheDirectory, workspaceSlug, projectSlug+".git")
}

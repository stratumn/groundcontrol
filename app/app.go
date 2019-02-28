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

	"groundcontrol/appcontext"
	"groundcontrol/gql"
	"groundcontrol/job"
	"groundcontrol/log"
	"groundcontrol/model"
	"groundcontrol/pubsub"
	"groundcontrol/relay"
	"groundcontrol/resolver"
	"groundcontrol/service"
	"groundcontrol/store"
	"groundcontrol/work"
)

// App starts Ground Control.
type App struct {
	sourcesFile             string
	keysFile                string
	listenAddress           string
	jobConcurrency          int
	logLevel                model.LogLevel
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

	waitGroup sync.WaitGroup
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
	appCtx := a.createModelContext()
	ctx, cancel := context.WithCancel(appcontext.With(ctx, appCtx))
	defer cancel()

	a.createBaseNodes(ctx)

	log := appCtx.Log
	systemID := appCtx.SystemID

	log.InfoWithOwner(ctx, systemID, "starting app")
	log.InfoWithOwner(
		ctx,
		systemID,
		"runtime %s %s %s",
		runtime.GOOS,
		runtime.GOARCH,
		runtime.Version(),
	)

	if err := a.createSources(ctx); err != nil {
		return err
	}

	if err := a.createKeys(ctx); err != nil {
		return err
	}

	if err := initHooks(ctx); err != nil {
		return err
	}

	router := chi.NewRouter()

	a.addCORS(router)
	a.addGQL(ctx, router)
	a.addPlayground(router)

	if a.ui != nil {
		a.addUI(router)
	}

	server := &http.Server{
		Addr:    a.listenAddress,
		Handler: router,
	}

	a.serve(ctx, server, cancel)

	if a.enableSignalHandling {
		a.handleSignals(ctx, server, cancel)
	}

	a.startJobs(ctx, cancel)
	a.startPeriodicJobs(ctx, cancel)

	if a.openBrowser && a.ui != nil {
		a.openAddressInBrowser(ctx)
	}

	<-ctx.Done()

	a.shutdown(ctx, server)

	return ctx.Err()
}

func (a *App) createModelContext() *appcontext.Context {
	return &appcontext.Context{
		Nodes:               store.NewMemory(),
		Log:                 log.NewLogger(a.logCap, a.logLevel),
		Jobs:                work.NewQueue(a.jobConcurrency),
		Services:            service.NewManager(),
		Subs:                pubsub.New(a.pubSubHistoryCap),
		GetGitSourcePath:    a.getGitSourcePath,
		GetProjectPath:      a.getProjectPath,
		GetProjectCachePath: a.getProjectCachePath,
		OpenEditorCommand:   a.openEditorCommand,
	}
}

func (a *App) createBaseNodes(ctx context.Context) {
	var (
		viewerID         = relay.EncodeID(model.NodeTypeUser)
		systemID         = relay.EncodeID(model.NodeTypeSystem)
		jobMetricsID     = relay.EncodeID(model.NodeTypeJobMetrics)
		serviceMetricsID = relay.EncodeID(model.NodeTypeServiceMetrics)
		logMetricsID     = relay.EncodeID(model.NodeTypeLogMetrics)
	)

	(&model.User{ID: viewerID}).MustStore(ctx)
	(&model.LogMetrics{ID: logMetricsID}).MustStore(ctx)
	(&model.JobMetrics{ID: jobMetricsID}).MustStore(ctx)
	(&model.ServiceMetrics{ID: serviceMetricsID}).MustStore(ctx)
	(&model.System{
		ID:               systemID,
		JobMetricsID:     jobMetricsID,
		LogMetricsID:     logMetricsID,
		ServiceMetricsID: serviceMetricsID,
	}).MustStore(ctx)

	appCtx := appcontext.Get(ctx)
	appCtx.ViewerID = viewerID
	appCtx.SystemID = systemID
}

func (a *App) createSources(ctx context.Context) error {
	appCtx := appcontext.Get(ctx)

	config, err := model.LoadSourcesConfigYAML(a.sourcesFile)
	if err != nil {
		return err
	}

	if err := config.Store(ctx); err != nil {
		return err
	}

	appCtx.Sources = config

	return nil
}

func (a *App) createKeys(ctx context.Context) error {
	appCtx := appcontext.Get(ctx)

	config, err := model.LoadKeysConfigYAML(a.keysFile)
	if err != nil {
		return err
	}

	if err := config.Store(ctx); err != nil {
		return err
	}

	appCtx.Keys = config

	return nil
}

func (a *App) addCORS(router *chi.Mux) {
	router.Use(cors.New(cors.Options{
		AllowCredentials: true,
		Debug:            false,
	}).Handler)
}

func (a *App) addGQL(ctx context.Context, router *chi.Mux) {
	appCtx := appcontext.Get(ctx)

	gqlConfig := gql.Config{
		Resolvers: &resolver.Resolver{AppCtx: appCtx},
	}

	gqlOptions := []handler.Option{
		handler.WebsocketUpgrader(websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		}),
		handler.ResolverMiddleware(modelContextMiddleware(appCtx)),
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
}

func (a *App) addPlayground(router *chi.Mux) {
	router.Handle("/graphql", handler.Playground("GraphQL playground", "/query"))
}

func (a *App) addUI(router *chi.Mux) {
	fileServer := http.FileServer(a.ui)
	router.NotFound(func(w http.ResponseWriter, req *http.Request) {
		if _, err := a.ui.Open(req.URL.Path); err != nil {
			// Rewrite other URLs to index for pushstate.
			req.URL.Path = ""
		}
		fileServer.ServeHTTP(w, req)
	})
}

func (a *App) serve(ctx context.Context, server *http.Server, cancel func()) {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	systemID := appCtx.SystemID

	a.waitGroup.Add(1)

	go func() {
		defer a.waitGroup.Done()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.ErrorWithOwner(
				ctx,
				systemID,
				"server crashed because %s",
				err.Error(),
			)
			cancel()
		}

		log.DebugWithOwner(ctx, systemID, "server terminated")
	}()

	log.DebugWithOwner(ctx, systemID, "starting server")

	log.InfoWithOwner(ctx, systemID, "app ready")
	if a.ui != nil {
		log.InfoWithOwner(ctx, systemID, "user interface on %s", a.listenAddress)
	}
	log.InfoWithOwner(
		ctx,
		systemID,
		"GraphQL playground on %s/graphql",
		a.listenAddress,
	)

}

func (a *App) handleSignals(ctx context.Context, server *http.Server, cancel func()) {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	systemID := appCtx.SystemID

	signalCh := make(chan os.Signal, 2)
	signal.Notify(signalCh, syscall.SIGTERM)
	signal.Notify(signalCh, syscall.SIGINT)

	go func() {
		log.DebugWithOwner(ctx, systemID, "received signal %d", <-signalCh)
		cancel()
	}()
}

func (a *App) startJobs(ctx context.Context, cancel func()) {
	appCtx := appcontext.Get(ctx)
	jobs := appCtx.Jobs
	log := appCtx.Log
	systemID := appCtx.SystemID

	a.waitGroup.Add(1)

	go func() {
		defer a.waitGroup.Done()
		log.DebugWithOwner(ctx, systemID, "starting jobs")

		if err := jobs.Work(ctx); err != nil && err != context.Canceled {
			log.ErrorWithOwner(
				ctx,
				systemID,
				"job manager crashed because %s",
				err.Error(),
			)
			cancel()
		}

		log.DebugWithOwner(ctx, systemID, "jobs terminated")
	}()
}

func (a *App) startPeriodicJobs(ctx context.Context, cancel func()) {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	systemID := appCtx.SystemID

	a.waitGroup.Add(1)

	go func() {
		defer a.waitGroup.Done()
		log.DebugWithOwner(ctx, systemID, "starting periodic jobs")

		err := job.StartPeriodic(
			ctx,
			a.periodicJobsInterval,
			func(ctx context.Context) []string {
				return job.LoadAllSources(ctx, false)
			},
			func(ctx context.Context) []string {
				return job.LoadAllCommits(ctx, false)
			},
		)
		if err != nil && err != context.Canceled {
			log.ErrorWithOwner(
				ctx,
				systemID,
				"periodic jobs crashed because %s",
				err.Error(),
			)
			cancel()
		}

		log.DebugWithOwner(ctx, systemID, "periodic jobs terminated")
	}()

}

func (a *App) shutdown(ctx context.Context, server *http.Server) {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	services := appCtx.Services
	systemID := appCtx.SystemID

	cleanCtx, cancel := context.WithTimeout(
		appcontext.With(context.Background(), appCtx),
		a.gracefulShutdownTimeout,
	)
	defer cancel()

	log.InfoWithOwner(ctx, systemID, "starting shutdown")

	a.waitGroup.Add(2)

	go func() {
		services.Clean(cleanCtx)
		a.waitGroup.Done()
	}()

	go func() {
		if err := server.Shutdown(cleanCtx); err != nil && err != context.Canceled {
			log.ErrorWithOwner(
				ctx,
				systemID,
				"server shutdown failed because %s",
				err.Error(),
			)
		}
		a.waitGroup.Done()
	}()

	doneCh := make(chan struct{})
	go func() {
		a.waitGroup.Wait()
		doneCh <- struct{}{}
	}()

	select {
	case <-cleanCtx.Done():
		log.ErrorWithOwner(
			ctx,
			systemID,
			"clean shutdown failed because %s",
			ctx.Err().Error(),
		)
		os.Exit(1)
	case <-doneCh:
		log.InfoWithOwner(ctx, systemID, "clean shutdown complete, goodbye!")
		os.Exit(0)
	}
}

func (a *App) openAddressInBrowser(ctx context.Context) {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	systemID := appCtx.SystemID

	addr, err := net.ResolveTCPAddr("tcp", a.listenAddress)
	if err != nil {
		log.WarningWithOwner(
			ctx,
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
			ctx,
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

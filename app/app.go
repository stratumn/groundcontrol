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
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/browser"

	"groundcontrol/appcontext"
	"groundcontrol/config"
	"groundcontrol/job"
	"groundcontrol/log"
	"groundcontrol/model"
	"groundcontrol/pubsub"
	"groundcontrol/relay"
	"groundcontrol/service"
	"groundcontrol/store"
	"groundcontrol/util"
	"groundcontrol/work"

	_ "net/http/pprof"
)

// App contains everything that's needed to start Ground Control.
type App struct {
	// All of these can be set by passing options to New().
	sourcesFile                   string
	keysFile                      string
	listenAddress                 string
	jobsConcurrency               int
	jobsChannelSize               int
	logLevel                      model.LogLevel
	logCap                        int
	pubSubHistoryCap              int
	subscriptionChannelSize       int
	periodicJobsInterval          time.Duration
	gracefulShutdownTimeout       time.Duration
	runnerGracefulShutdownTimeout time.Duration
	ui                            http.FileSystem
	openBrowser                   bool
	gitSourcesDirectory           string
	workspacesDirectory           string
	cacheDirectory                string
	openEditorCommand             string
	enableApolloTracing           bool
	enableSignalHandling          bool
	pprofListenAddress            string
	newRunner                     appcontext.NewRunner
	// The app needs to launch a few Goroutines, and a wait group is used to
	// make sure they finish before exiting.
	// TODO: see if golang.org/x/sync/errgroup is a better option
	waitGroup sync.WaitGroup
}

// New creates a new App with given options. All options have default values.
func New(opts ...Opt) *App {
	app := &App{
		sourcesFile:                   DefaultSourcesFile,
		keysFile:                      DefaultKeysFile,
		listenAddress:                 DefaultListenAddress,
		jobsConcurrency:               DefaultJobsConcurrency,
		jobsChannelSize:               DefaultJobsChannelSize,
		logLevel:                      DefaultLogLevel,
		logCap:                        DefaultLogCap,
		pubSubHistoryCap:              DefaultPubSubHistoryCap,
		subscriptionChannelSize:       DefaultSubscriptionChannelSize,
		periodicJobsInterval:          DefaultPeriodicJobsInterval,
		gracefulShutdownTimeout:       DefaultGracefulShutdownTimeout,
		openBrowser:                   DefaultOpenBrowser,
		gitSourcesDirectory:           DefaultGitSourcesDirectory,
		workspacesDirectory:           DefaultWorkspacesDirectory,
		cacheDirectory:                DefaultCacheDirectory,
		newRunner:                     DefaultNewRunner,
		runnerGracefulShutdownTimeout: DefaultRunnerGracefulShutdownTimeout,
		openEditorCommand:             DefaultOpenEditorCommand,
		enableApolloTracing:           DefaultEnableApolloTracing,
		enableSignalHandling:          DefaultEnableSignalHandling,
		pprofListenAddress:            DefaultPprofListenAddress,
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

// Start starts the app. It blocks until an error occurs, the context is
// canceled, or an exit signal is received. It will do some cleanup before
// returning, which can take some time.
func (a *App) Start(ctx context.Context) error {
	// Augment the context with an appcontext.Context to propagate variables
	// to app functions.
	appCtx := a.createAppContext()
	// When an exit signal is received, or one of the Goroutines returns,
	// cancel() is called to initiate a shutdown.
	ctx, cancel := context.WithCancel(appcontext.With(ctx, appCtx))
	defer cancel()
	a.createBaseNodes(ctx) // sets appCtx.systemID
	appCtx.Log.InfoWithOwner(ctx, appCtx.SystemID, "starting app")
	if err := a.createSources(ctx); err != nil {
		return err
	}
	if err := a.createKeys(ctx); err != nil {
		return err
	}
	if err := initHooks(ctx); err != nil {
		return err
	}
	// Start the HTTP server as soon as possible to reduce the risk of the
	// browser opening before the UI is ready to be served.
	server := a.createServer(ctx)
	a.serve(ctx, server, cancel)
	a.startJobs(ctx, cancel)
	a.startPeriodicJobs(ctx, cancel)
	if a.enableSignalHandling {
		a.handleSignals(ctx, server, cancel)
	}
	if a.openBrowser && a.ui != nil {
		a.openUI(ctx)
	}
	if a.pprofListenAddress != "" {
		a.startPprof(ctx)
	}
	appCtx.Log.InfoWithOwner(ctx, appCtx.SystemID, "app ready")
	// Block until the context is canceled by either the app or the parent
	// context, then shutdown everything.
	<-ctx.Done()
	a.shutdown(ctx, server)
	return ctx.Err()
}

// createAppContext creates the app context that will be attached to a Go
// context. Functions in the program can retrieve it by calling
// appcontext.Get().
func (a *App) createAppContext() *appcontext.Context {
	return &appcontext.Context{
		Nodes:                         store.NewMemory(),
		Log:                           log.NewLogger(a.logCap, a.logLevel),
		Jobs:                          work.NewQueue(a.jobsConcurrency, a.jobsChannelSize),
		Services:                      service.NewManager(),
		Subs:                          pubsub.New(a.pubSubHistoryCap),
		SubChannelSize:                a.subscriptionChannelSize,
		GetGitSourcePath:              a.getGitSourcePath,
		GetProjectPath:                a.getProjectPath,
		GetProjectCachePath:           a.getProjectCachePath,
		NewRunner:                     a.newRunner,
		RunnerGracefulShutdownTimeout: a.runnerGracefulShutdownTimeout,
		OpenEditorCommand:             a.openEditorCommand,
	}
}

// createBaseNodes creates the Relay nodes that are needed before other nodes
// can be created.
func (a *App) createBaseNodes(ctx context.Context) {
	var (
		// Since the nodes are unique, their type is enough to create a unique
		// ID, for as long as there is only one user and system.
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
	// The viewer and system IDs are stored in the app context and are the only
	// IDs that are globally needed since all other nodes are children.
	appCtx := appcontext.Get(ctx)
	appCtx.ViewerID = viewerID
	appCtx.SystemID = systemID
}

// createSources loads the sources config file and creates the Relay nodes for
// them.
func (a *App) createSources(ctx context.Context) error {
	config, err := model.LoadSourcesConfigYAML(a.sourcesFile)
	if err != nil {
		return err
	}
	if err := config.Store(ctx); err != nil {
		return err
	}
	appcontext.Get(ctx).Sources = config
	return nil
}

// createKeys loads the keys config file and creates the Relay nodes for them.
func (a *App) createKeys(ctx context.Context) error {
	cfg, err := config.LoadKeysYAML(a.keysFile)
	if err != nil {
		return err
	}
	appcontext.Get(ctx).Keys = cfg
	model.InjectKeysConfig(ctx)
	return nil
}

// createServer create the HTTP server.
func (a *App) createServer(ctx context.Context) *http.Server {
	r := newRouter()
	// Middlewares need to be added before routes.
	r.EnableCORS()
	r.EnableGQL(appcontext.Get(ctx), a.enableApolloTracing)
	r.EnablePlayground()
	if a.ui != nil {
		r.EnableUI(a.ui)
	}
	return &http.Server{Addr: a.listenAddress, Handler: r}
}

// serve starts the HTTP server in a Goroutine.
func (a *App) serve(ctx context.Context, server *http.Server, cancel func()) {
	a.proc(ctx, "http server", cancel, func(ctx context.Context) error {
		return server.ListenAndServe()
	})
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	systemID := appCtx.SystemID
	if a.ui != nil {
		log.InfoWithOwner(ctx, systemID, "user interface on %s", a.listenAddress)
	}
	log.InfoWithOwner(ctx, systemID, "GraphQL playground on %s/graphql", a.listenAddress)
}

// startJobs starts the work queue in a Goroutine.
func (a *App) startJobs(ctx context.Context, cancel func()) {
	a.proc(ctx, "work queue", cancel, appcontext.Get(ctx).Jobs.Work)
}

// startPeriodicJobs starts peridically creating jobs to synchronize sources
// and workspaces in a Goroutine.
func (a *App) startPeriodicJobs(ctx context.Context, cancel func()) {
	a.proc(ctx, "periodic jobs", cancel, func(ctx context.Context) error {
		return job.StartPeriodic(
			ctx,
			a.periodicJobsInterval,
			func(ctx context.Context) []string {
				return job.SyncSources(ctx, false)
			},
			func(ctx context.Context) []string {
				return job.SyncWorkspaces(ctx, false)
			},
		)
	})
}

// handleSignals starts a Goroutine that will cancel the app context once
// a SIGTERM or SIGINT signal is received.
func (a *App) handleSignals(ctx context.Context, server *http.Server, cancel func()) {
	signalCh := make(chan os.Signal, 2)
	signal.Notify(signalCh, syscall.SIGTERM)
	signal.Notify(signalCh, syscall.SIGINT)
	go func() {
		appCtx := appcontext.Get(ctx)
		sig := <-signalCh
		appCtx.Log.DebugWithOwner(ctx, appCtx.SystemID, "received signal %d", sig)
		cancel()
	}()
}

// shutdown terminates all the running Goroutines, and cleans up the app before
// exiting the process.
// TODO: for testing there needs to be a way not to exit the process.
func (a *App) shutdown(ctx context.Context, server *http.Server) {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	systemID := appCtx.SystemID
	log.InfoWithOwner(ctx, systemID, "starting shutdown")
	// At this point the main context is already done, so a new one is created
	// with a timeout to allow the remaining Goroutines to clean up.
	cleanCtx, cancel := context.WithTimeout(
		appcontext.With(context.Background(), appCtx),
		a.gracefulShutdownTimeout,
	)
	defer cancel()

	a.waitGroup.Add(2)
	go func() {
		// The service manager doesn't have its own Goroutine, but it needs to
		// stop all running services.
		appCtx.Services.Clean(cleanCtx)
		a.waitGroup.Done()
	}()
	go func() {
		// Even though it's not important, try to cleanly shutdown the HTTP
		// server.
		if err := server.Shutdown(cleanCtx); err != nil && err != context.Canceled {
			log.ErrorWithOwner(ctx, systemID, "http server shutdown failed because %s", err.Error())
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
		// Not all Goroutine exited in time, so exit the process with an error.
		log.ErrorWithOwner(ctx, systemID, "graceful shutdown failed because %s", ctx.Err().Error())
		os.Exit(1)
	case <-doneCh:
		// All Goroutines stopped in time.
		// TODO: if a Goroutine exited with an unexpected error, it should exit
		// with status 1.
		log.InfoWithOwner(ctx, systemID, "graceful shutdown complete, goodbye!")
		os.Exit(0)
	}
}

// openUI opens the URL of the user interface in the default browser.
func (a *App) openUI(ctx context.Context) {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	systemID := appCtx.SystemID
	url, err := util.AddressURL(a.listenAddress)
	if err != nil {
		log.WarningWithOwner(ctx, systemID, "could not resolve UI address because %s", err.Error())
		return
	}
	if err := browser.OpenURL(url); err != nil {
		log.WarningWithOwner(ctx, systemID, "could not open UI in browser because %s", err.Error())
	}
}

// startPprof start another HTTP server for the Go profiler, which can be used
// to debug CPU and memory usage, amongst other things.
func (a *App) startPprof(ctx context.Context) {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	systemID := appCtx.SystemID
	log.DebugWithOwner(ctx, systemID, "starting pprof")

	go func() {
		if err := http.ListenAndServe(a.pprofListenAddress, nil); err != nil && err != http.ErrServerClosed {
			log.ErrorWithOwner(ctx, systemID, "pprof crashed because %s", err.Error())
		}
		log.DebugWithOwner(ctx, systemID, "pprof terminated")
	}()
}

// getGitSourcePath returns the path to the directory where the files of a Git
// source are stored.
func (a *App) getGitSourcePath(repo, reference string) string {
	name := path.Base(repo)
	ext := path.Ext(name)
	name = name[:len(name)-len(ext)]
	return filepath.Join(a.gitSourcesDirectory, name, reference)
}

// getProjectPath returns the path to the directory where the files of a
// project are stored.
func (a *App) getProjectPath(workspaceSlug, projectSlug string) string {
	return filepath.Join(a.workspacesDirectory, workspaceSlug, projectSlug)
}

// getProjectCachePath returns the path to the directory where the files of a
// project cache are stored. The cache of a project is just a bare clone of the
// repository.
func (a *App) getProjectCachePath(workspaceSlug, projectSlug string) string {
	return filepath.Join(a.cacheDirectory, workspaceSlug, projectSlug+".git")
}

// proc is used to launch a long-running Goroutine, and takes care of updating
// the wait group.
func (a *App) proc(ctx context.Context, name string, cancel context.CancelFunc, fn func(ctx context.Context) error) {
	appCtx := appcontext.Get(ctx)
	log := appCtx.Log
	systemID := appCtx.SystemID
	log.InfoWithOwner(ctx, systemID, "starting %s", name)
	a.waitGroup.Add(1)
	go func() {
		defer a.waitGroup.Done()
		err := fn(ctx)
		if err != nil && err != context.Canceled && err != http.ErrServerClosed {
			log.ErrorWithOwner(ctx, systemID, "%s crashed because %s", name, err.Error())
			cancel()
		}
		log.InfoWithOwner(ctx, systemID, "%s stopped", name)
	}()
}

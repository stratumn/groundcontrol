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
	"log"
	"net/http"
	"path/filepath"
	"time"

	homedir "github.com/mitchellh/go-homedir"

	"groundcontrol/model"
)

const (
	// DefaultListenAddress is the default listen address.
	DefaultListenAddress = "localhost:4444"
	// DefaultJobsConcurrency is the default jobs concurrency.
	DefaultJobsConcurrency = 2
	// DefaultJobsChannelSize is the default jobs channel size.
	DefaultJobsChannelSize = 1024
	// DefaultLogLevel is the default log level.
	DefaultLogLevel = model.LogLevelInfo
	// DefaultLogCap is the default capacity of the logger.
	DefaultLogCap = 10000
	// DefaultPubSubHistoryCap is the default capacity of the PubSub history.
	DefaultPubSubHistoryCap = 20
	// DefaultSubscriptionChannelSize is the default subscription channel size.
	DefaultSubscriptionChannelSize = 1024
	// DefaultPeriodicJobsInterval is the default periodic jobs interval.
	DefaultPeriodicJobsInterval = time.Minute
	// DefaultGracefulShutdownTimeout is the default graceful shutdown timeout.
	DefaultGracefulShutdownTimeout = 20 * time.Second
	// DefaultOpenBrowser is whether to open the user interface in a browser by default.
	DefaultOpenBrowser = true
	// DefaultEnableApolloTracing is whether to enable Apollo tracing by default.
	DefaultEnableApolloTracing = false
	// DefaultEnableSignalHandling is whether to enable signal handling by default.
	DefaultEnableSignalHandling = true
	// DefaultOpenEditorCommand is the default open editor command.
	DefaultOpenEditorCommand = "code --goto %s"
	// DefaultPprofListenAddress is the default pprof listen address.
	DefaultPprofListenAddress = ""
)

var (
	// DefaultSettingsFile is the default settings file.
	DefaultSettingsFile = "settings.yml"
	// DefaultSourcesFile is the default sources file.
	DefaultSourcesFile = "sources.yml"
	// DefaultKeysFile is the default keys file.
	DefaultKeysFile = "keys.yml"
	// DefaultGitSourcesDirectory is the default Git sources directory.
	DefaultGitSourcesDirectory = "git-sources"
	// DefaultWorkspacesDirectory is the default workspace directory.
	DefaultWorkspacesDirectory = "workspaces"
	// DefaultCacheDirectory is the default cache directory.
	DefaultCacheDirectory = "cache"
)

func init() {
	home, err := homedir.Dir()
	if err != nil {
		log.Printf("WARNING\tcould not resolve home directory because %s", err.Error())
		return
	}
	DefaultSettingsFile = filepath.Join(home, "groundcontrol", DefaultSettingsFile)
	DefaultSourcesFile = filepath.Join(home, "groundcontrol", DefaultSourcesFile)
	DefaultKeysFile = filepath.Join(home, "groundcontrol", DefaultKeysFile)
	DefaultGitSourcesDirectory = filepath.Join(home, "groundcontrol", DefaultGitSourcesDirectory)
	DefaultWorkspacesDirectory = filepath.Join(home, "groundcontrol", DefaultWorkspacesDirectory)
	DefaultCacheDirectory = filepath.Join(home, "groundcontrol", DefaultCacheDirectory)
}

// Opt represents an app option.
type Opt func(*App)

// OptSourcesFile sets the sources file.
func OptSourcesFile(filename string) Opt {
	return func(app *App) {
		app.sourcesFile = filename
	}
}

// OptKeysFile sets the keys file.
func OptKeysFile(filename string) Opt {
	return func(app *App) {
		app.keysFile = filename
	}
}

// OptListenAddress sets the listen address.
func OptListenAddress(address string) Opt {
	return func(app *App) {
		app.listenAddress = address
	}
}

// OptJobsConcurrency sets the concurrency of the work queue.
func OptJobsConcurrency(concurrency int) Opt {
	return func(app *App) {
		app.jobsConcurrency = concurrency
	}
}

// OptJobsChannelSize sets the size of a priority channel in the work queue.
func OptJobsChannelSize(size int) Opt {
	return func(app *App) {
		app.jobsChannelSize = size
	}
}

// OptLogLevel sets the minimum level for log messages.
func OptLogLevel(level model.LogLevel) Opt {
	return func(app *App) {
		app.logLevel = level
	}
}

// OptPubSubHistoryCap sets the capacity of the PubSub history cap.
func OptPubSubHistoryCap(cap int) Opt {
	return func(app *App) {
		app.pubSubHistoryCap = cap
	}
}

// OptSubscriptionChannelSize sets the size of a subscription channel.
func OptSubscriptionChannelSize(size int) Opt {
	return func(app *App) {
		app.subscriptionChannelSize = size
	}
}

// OptLogCap sets the capacity of the logger.
func OptLogCap(cap int) Opt {
	return func(app *App) {
		app.logCap = cap
	}
}

// OptPeriodicJobsInterval sets the time to wait between periodic job.
func OptPeriodicJobsInterval(interval time.Duration) Opt {
	return func(app *App) {
		app.periodicJobsInterval = interval
	}
}

// OptGracefulShutdownTimeout sets the maximum duration for a graceful shutdown.
func OptGracefulShutdownTimeout(timeout time.Duration) Opt {
	return func(app *App) {
		app.gracefulShutdownTimeout = timeout
	}
}

// OptUI sets the file system for the UI.
func OptUI(fs http.FileSystem) Opt {
	return func(app *App) {
		app.ui = fs
	}
}

// OptOpenBrowser tells the app whether to open the user interface in a browser.
func OptOpenBrowser(open bool) Opt {
	return func(app *App) {
		app.openBrowser = open
	}
}

// OptGitSourcesDirectory sets the directory for Git sources.
func OptGitSourcesDirectory(dir string) Opt {
	return func(app *App) {
		app.gitSourcesDirectory = dir
	}
}

// OptWorkspacesDirectory sets the directory for workspaces.
func OptWorkspacesDirectory(dir string) Opt {
	return func(app *App) {
		app.workspacesDirectory = dir
	}
}

// OptCacheDirectory sets the directory for the cache.
func OptCacheDirectory(dir string) Opt {
	return func(app *App) {
		app.cacheDirectory = dir
	}
}

// OptOpenEditorCommand sets the command to use to open an editor.
func OptOpenEditorCommand(command string) Opt {
	return func(app *App) {
		app.openEditorCommand = command
	}
}

// OptEnableApolloTracing tells the app whether to enable the Apollo tracing middleware.
func OptEnableApolloTracing(enable bool) Opt {
	return func(app *App) {
		app.enableApolloTracing = enable
	}
}

// OptEnableSignalHandling tells the app whether to handle exit signals.
func OptEnableSignalHandling(enable bool) Opt {
	return func(app *App) {
		app.enableSignalHandling = enable
	}
}

// OptPprofListenAddress sets the pprof listen address.
func OptPprofListenAddress(address string) Opt {
	return func(app *App) {
		app.pprofListenAddress = address
	}
}

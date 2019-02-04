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

	"github.com/stratumn/groundcontrol/models"
)

const (
	// DefaultConfigFilename is the default filename of the config file.
	DefaultConfigFilename = "groundcontrol.yml"

	// DefaultListenAddress is the default listen address.
	DefaultListenAddress = ":3333"

	// DefaultJobConcurrency is the default concurrency of the job manager.
	DefaultJobConcurrency = 2

	// DefaultLogLevel is the default log level.
	DefaultLogLevel = models.LogLevelInfo

	// DefaultLogCap is the default capacity of the logger.
	DefaultLogCap = 10000

	// DefaultCheckSourcesInterval is the default check sources interval.
	DefaultCheckSourcesInterval = time.Minute

	// DefaultCheckProjectsInterval is the default check projects interval.
	DefaultCheckProjectsInterval = time.Minute

	// DefaultGracefulShutdownTimeout is the default graceful shutdown timeout.
	DefaultGracefulShutdownTimeout = 20 * time.Second

	// DefaultOpenBrowser is whether to open the user interface in a browser by default.
	DefaultOpenBrowser = true

	// DefaultEnableApolloTracing is whether to enable Apollo tracing by default.
	DefaultEnableApolloTracing = false

	// DefaultEnableSignalHandling is whether to enable signal handling by default.
	DefaultEnableSignalHandling = true
)

var (
	// DefaultSettingsFile is the default settings file.
	DefaultSettingsFile = "settings.yml"

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
	DefaultWorkspacesDirectory = filepath.Join(home, "groundcontrol", DefaultWorkspacesDirectory)
	DefaultCacheDirectory = filepath.Join(home, "groundcontrol", DefaultCacheDirectory)
}

// Opt represents an app option.
type Opt func(*App)

// OptConfigFilenames adds config files. This option can be added multiple times.
func OptConfigFilenames(filenames ...string) Opt {
	return func(app *App) {
		app.configFilenames = append(app.configFilenames, filenames...)
	}
}

// OptListenAddress sets the listen address.
func OptListenAddress(address string) Opt {
	return func(app *App) {
		app.listenAddress = address
	}
}

// OptJobConcurrency sets the concurrency of the job manager.
func OptJobConcurrency(concurrency int) Opt {
	return func(app *App) {
		app.jobConcurrency = concurrency
	}
}

// OptLogLevel sets the minimum level for log messages.
func OptLogLevel(level models.LogLevel) Opt {
	return func(app *App) {
		app.logLevel = level
	}
}

// OptLogCap sets the capacity of the logger.
func OptLogCap(cap int) Opt {
	return func(app *App) {
		app.logCap = cap
	}
}

// OptCheckSourcesInterval sets the time to wait between periodic jobs used to check the state of sources.
func OptCheckSourcesInterval(interval time.Duration) Opt {
	return func(app *App) {
		app.checkSourcesInterval = interval
	}
}

// OptCheckProjectsInterval sets the time to wait between periodic jobs used to check the state of projects.
func OptCheckProjectsInterval(interval time.Duration) Opt {
	return func(app *App) {
		app.checkProjectsInterval = interval
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

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
	"path"
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

	// DefaultCheckProjectsInterval is the default check projects interval.
	DefaultCheckProjectsInterval = time.Minute

	// DefaultGracefulShutdownTimeout is the default graceful shutdown timeout.
	DefaultGracefulShutdownTimeout = 20 * time.Second

	// DefaultOpenBrowser is whether to open the user interface in a browser by default.
	DefaultOpenBrowser = true
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

	DefaultSettingsFile = filepath.Join(home, "groundcontrol", "settings.yml")
	DefaultWorkspacesDirectory = filepath.Join(home, "groundcontrol", "workspaces")
	DefaultCacheDirectory = filepath.Join(home, "groundcontrol", "cache")
}

// DefaultProjectPathGetter is the default ProjectPathGetter.
func DefaultProjectPathGetter(workspaceSlug, repo, branch string) string {
	name := path.Base(repo)
	ext := path.Ext(name)
	name = name[:len(name)-len(ext)]
	return filepath.Join("workspaces", workspaceSlug, name)
}

// DefaultProjectCachePathGetter is the default ProjectCachePathGetter.
func DefaultProjectCachePathGetter(workspaceSlug, repo, branch string) string {
	name := path.Base(repo)
	ext := path.Ext(name)
	name = name[:len(name)-len(ext)]
	return filepath.Join("cache", workspaceSlug, name+".git")
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

// OptCheckProjectsInterval sets the time to wait between periodic jobs used to check the state of projects.
func OptCheckProjectsInterval(interval time.Duration) Opt {
	return func(app *App) {
		app.checkProjectsInterval = interval
	}
}

// OptDisableSignalHandling tells the app not to listen to exit signals.
func OptDisableSignalHandling() Opt {
	return func(app *App) {
		app.disableSignalHandling = true
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

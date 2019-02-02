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

// Package cmd contains the commands for the app.
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stratumn/groundcontrol/app"
	"github.com/stratumn/groundcontrol/models"
)

var (
	userInterface http.FileSystem

	settingsFile            string
	listenAddress           string
	jobConcurrency          int
	logLevel                string
	logCap                  int
	checkProjectsInterval   time.Duration
	gracefulShutdownTimeout time.Duration
	openBrowser             bool
	workspacesDirectory     string
	cacheDirectory          string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "groundcontrol [groundcontrol.yml ...]",
	Short: "Ground Control is an app to manage multiple Git repositories",
	Long: `Ground Control is an application to help deal with multi-repository development using a user friendly web interface.

Complete documentation is available at https://github.com/stratumn/groundcontrol.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		app := app.New(
			app.OptConfigFilenames(args...),
			app.OptListenAddress(viper.GetString("listen-address")),
			app.OptJobConcurrency(viper.GetInt("job-concurrency")),
			app.OptLogLevel(models.LogLevel(strings.ToUpper(viper.GetString("log-level")))),
			app.OptLogCap(viper.GetInt("log-cap")),
			app.OptCheckProjectsInterval(viper.GetDuration("check-projects-interval")),
			app.OptGracefulShutdownTimeout(viper.GetDuration("graceful-shutdown-timeout")),
			app.OptOpenBrowser(viper.GetBool("open-browser")),
			app.OptWorkspacesDirectory(viper.GetString("workspaces-directory")),
			app.OptCacheDirectory(viper.GetString("cache-directory")),
			app.OptUI(userInterface),
		)

		return app.Start(ctx)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ui http.FileSystem) {
	userInterface = ui

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initSettings)

	rootCmd.PersistentFlags().StringVar(&settingsFile, "settings", app.DefaultSettingsFile, "settings file")
	rootCmd.PersistentFlags().StringVar(&listenAddress, "listen-address", app.DefaultListenAddress, "address the server should listen on")
	rootCmd.PersistentFlags().IntVar(&jobConcurrency, "job-concurrency", app.DefaultJobConcurrency, "how many jobs can run concurrency")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", app.DefaultLogLevel.String(), "minimum level of log messages (debug, info, warning, error)")
	rootCmd.PersistentFlags().IntVar(&logCap, "log-cap", app.DefaultLogCap, "maximum number of messages the logger will keep")
	rootCmd.PersistentFlags().DurationVar(&checkProjectsInterval, "check-projects-interval", app.DefaultCheckProjectsInterval, "how often to check if projects have changed")
	rootCmd.PersistentFlags().DurationVar(&gracefulShutdownTimeout, "graceful-shutdown-timeout", app.DefaultGracefulShutdownTimeout, "maximum amount of time allowed to gracefully shutdown the app")
	rootCmd.PersistentFlags().BoolVar(&openBrowser, "open-browser", app.DefaultOpenBrowser, "open the user interface in a browser")
	rootCmd.PersistentFlags().StringVar(&workspacesDirectory, "workspaces-directory", app.DefaultWorkspacesDirectory, "directory for workspaces")
	rootCmd.PersistentFlags().StringVar(&cacheDirectory, "cache-directory", app.DefaultCacheDirectory, "directory for the cache")

	for _, flagName := range []string{
		"listen-address",
		"job-concurrency",
		"log-level",
		"log-cap",
		"check-projects-interval",
		"graceful-shutdown-timeout",
		"open-browser",
		"workspaces-directory",
		"cache-directory",
	} {
		viper.BindPFlag(flagName, rootCmd.PersistentFlags().Lookup(flagName))
	}
}

// initSettings reads in settings file and ENV variables if set.
func initSettings() {
	if settingsFile != "" {
		// Use settings file from the flag.
		viper.SetConfigFile(settingsFile)
	} else {
		// Search settings in home directory with name ".test" (without extension).
		viper.AddConfigPath(filepath.Dir(app.DefaultSettingsFile))
		viper.SetConfigName(filepath.Base(app.DefaultSettingsFile))
		viper.SetConfigType(filepath.Ext(app.DefaultSettingsFile))
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a settings file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using settings file:", viper.ConfigFileUsed())
	}
}

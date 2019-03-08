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
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"groundcontrol/app"
	"groundcontrol/model"
)

var (
	settingsFile  string
	userInterface http.FileSystem
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "groundcontrol",
	Short: "Ground Control is an app to manage multiple Git repositories",
	Long: `Ground Control is an application to help deal with multi-repository development using a user friendly web interface.

Complete documentation is available at https://github.com/stratumn/groundcontrol.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app := app.New(
			app.OptSourcesFile(viper.GetString("sources-file")),
			app.OptKeysFile(viper.GetString("keys-file")),
			app.OptListenAddress(viper.GetString("listen-address")),
			app.OptJobsConcurrency(viper.GetInt("jobs-concurrency")),
			app.OptJobsChannelSize(viper.GetInt("jobs-channel-size")),
			app.OptLogLevel(model.LogLevel(strings.ToUpper(viper.GetString("log-level")))),
			app.OptLogCap(viper.GetInt("log-cap")),
			app.OptPubSubHistoryCap(viper.GetInt("pubsub-history-cap")),
			app.OptSubscriptionChannelSize(viper.GetInt("subscription-channel-size")),
			app.OptPeriodicJobsInterval(viper.GetDuration("periodic-jobs-interval")),
			app.OptGracefulShutdownTimeout(viper.GetDuration("graceful-shutdown-timeout")),
			app.OptOpenBrowser(viper.GetBool("open-browser")),
			app.OptGitSourcesDirectory(viper.GetString("git-sources-directory")),
			app.OptWorkspacesDirectory(viper.GetString("workspaces-directory")),
			app.OptCacheDirectory(viper.GetString("cache-directory")),
			app.OptOpenEditorCommand(viper.GetString("open-editor-command")),
			app.OptEnableApolloTracing(viper.GetBool("enable-apollo-tracing")),
			app.OptPprofListenAddress(viper.GetString("pprof-listen-address")),
			app.OptUI(userInterface),
		)
		return app.Start(context.Background())
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(ui http.FileSystem) {
	userInterface = ui
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initSettings)
	rootCmd.PersistentFlags().StringVar(&settingsFile, "settings-file", app.DefaultSettingsFile, "settings file")
	rootCmd.PersistentFlags().String("sources-file", app.DefaultSourcesFile, "sources config file")
	rootCmd.PersistentFlags().String("keys-file", app.DefaultKeysFile, "keys config file")
	rootCmd.PersistentFlags().String("listen-address", app.DefaultListenAddress, "address the server should listen on")
	rootCmd.PersistentFlags().Int("jobs-concurrency", app.DefaultJobsConcurrency, "how many jobs can run concurrency")
	rootCmd.PersistentFlags().Int("jobs-channel-size", app.DefaultJobsChannelSize, "how many jobs a work queue can hold")
	rootCmd.PersistentFlags().String("log-level", app.DefaultLogLevel.String(), "minimum level of log messages (debug, info, warning, error)")
	rootCmd.PersistentFlags().Int("log-cap", app.DefaultLogCap, "maximum number of messages the logger will keep")
	rootCmd.PersistentFlags().Int("pubsub-history-cap", app.DefaultLogCap, "maximum number of messages the subscription manager will keep")
	rootCmd.PersistentFlags().Int("subscription-channel-size", app.DefaultSubscriptionChannelSize, "how many messages a subscription channel can hold")
	rootCmd.PersistentFlags().Duration("periodic-jobs-interval", app.DefaultPeriodicJobsInterval, "how long to wait between rounds of periodic jobs")
	rootCmd.PersistentFlags().Duration("graceful-shutdown-timeout", app.DefaultGracefulShutdownTimeout, "maximum amount of time allowed to gracefully shutdown the app")
	rootCmd.PersistentFlags().Bool("open-browser", app.DefaultOpenBrowser, "open the user interface in a browser")
	rootCmd.PersistentFlags().String("git-sources-directory", app.DefaultGitSourcesDirectory, "directory for Git sources")
	rootCmd.PersistentFlags().String("workspaces-directory", app.DefaultWorkspacesDirectory, "directory for workspaces")
	rootCmd.PersistentFlags().String("cache-directory", app.DefaultCacheDirectory, "directory for the cache")
	rootCmd.PersistentFlags().String("open-editor-command", app.DefaultOpenEditorCommand, "command issued to open a text editor")
	rootCmd.PersistentFlags().Bool("enable-apollo-tracing", app.DefaultEnableApolloTracing, "enable the Apollo tracing middleware")
	rootCmd.PersistentFlags().String("pprof-listen-address", app.DefaultPprofListenAddress, "address the profiler should listen on")
	for _, flagName := range []string{
		"sources-file",
		"keys-file",
		"listen-address",
		"jobs-concurrency",
		"jobs-channel-size",
		"log-level",
		"log-cap",
		"pubsub-history-cap",
		"subscription-channel-size",
		"periodic-jobs-interval",
		"graceful-shutdown-timeout",
		"open-browser",
		"git-sources-directory",
		"workspaces-directory",
		"cache-directory",
		"open-editor-command",
		"enable-apollo-tracing",
		"pprof-listen-address",
	} {
		if err := viper.BindPFlag(flagName, rootCmd.PersistentFlags().Lookup(flagName)); err != nil {
			panic(err)
		}
	}
}

// initSettings reads in settings file and ENV variables if set.
func initSettings() {
	if settingsFile != "" {
		viper.SetConfigFile(settingsFile)
	} else {
		viper.AddConfigPath(filepath.Dir(app.DefaultSettingsFile))
		viper.SetConfigName(filepath.Base(app.DefaultSettingsFile))
		viper.SetConfigType(filepath.Ext(app.DefaultSettingsFile))
	}
	viper.SetEnvPrefix("groundcontrol")
	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		log.Println("INFO\tusing settings file", viper.ConfigFileUsed())
	}
}

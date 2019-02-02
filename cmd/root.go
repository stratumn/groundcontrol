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

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stratumn/groundcontrol/app"
	"github.com/stratumn/groundcontrol/models"
)

var userInterface http.FileSystem
var settingsFile string

var (
	listenAddress           string
	jobConcurrency          int
	logLevel                string
	logCap                  int
	checkProjectsInterval   time.Duration
	gracefulShutdownTimeout time.Duration
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
			app.OptListenAddress(listenAddress),
			app.OptJobConcurrency(jobConcurrency),
			app.OptLogLevel(models.LogLevel(strings.ToUpper(logLevel))),
			app.OptLogCap(logCap),
			app.OptCheckProjectsInterval(checkProjectsInterval),
			app.OptGracefulShutdownTimeout(gracefulShutdownTimeout),
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

	rootCmd.PersistentFlags().StringVar(&settingsFile, "settings", "", "settings file (default is $HOME/.groundcontrol/settings.yml)")
	rootCmd.PersistentFlags().StringVar(&listenAddress, "listen-address", app.DefaultListenAddress, "address the server should listen on")
	rootCmd.PersistentFlags().IntVar(&jobConcurrency, "job-concurrency", app.DefaultJobConcurrency, "how many jobs can run concurrency")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", app.DefaultLogLevel.String(), "minimum level of log messages (debug, info, warning, error)")
	rootCmd.PersistentFlags().IntVar(&logCap, "log-cap", app.DefaultLogCap, "maximum number of messages the logger will keep")
	rootCmd.PersistentFlags().DurationVar(&checkProjectsInterval, "check-projects-interval", app.DefaultCheckProjectsInterval, "how often to check if projects have changed")
	rootCmd.PersistentFlags().DurationVar(&gracefulShutdownTimeout, "graceful-shutdown-timeout", app.DefaultGracefulShutdownTimeout, "maximum amount of time allowed to gracefully shutdown the app")
}

// initSettings reads in settings file and ENV variables if set.
func initSettings() {
	if settingsFile != "" {
		// Use settings file from the flag.
		viper.SetConfigFile(settingsFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search settings in home directory with name ".test" (without extension).
		viper.AddConfigPath(filepath.Join(home, "groundcontrol"))
		viper.SetConfigName("settings")
		viper.SetConfigType("yml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a settings file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using settings file:", viper.ConfigFileUsed())
	}
}

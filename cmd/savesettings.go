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

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"groundcontrol/app"
)

// saveSettingsCmd represents the save-settings command.
var saveSettingsCmd = &cobra.Command{
	Use:   fmt.Sprintf("save-settings [%s]", app.DefaultSettingsFile),
	Short: "Save settings to a file",
	Long: `Save settings to a file to avoid having to specify flags when Ground Control is launched.

It will overwrite the file if it exists.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return viper.WriteConfigAs(args[0])
		}
		return viper.WriteConfigAs(app.DefaultSettingsFile)
	},
}

func init() {
	rootCmd.AddCommand(saveSettingsCmd)
}

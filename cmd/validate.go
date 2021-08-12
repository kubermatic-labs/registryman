/*
   Copyright 2021 The Kubermatic Kubernetes Platform contributors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package cmd

import (
	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration files",
	Long:  `Validate the configuration files`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config.SetLogger(logger)
		logger.Info("validating config files", "dir", args[0])
		_, err := config.ReadLocalManifests(args[0], nil)
		if err == nil {
			logger.Info("config files are valid")
		} else {
			logger.V(-1).Info("config files are not valid",
				"error", err.Error())
			return err
		}
		return nil

	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

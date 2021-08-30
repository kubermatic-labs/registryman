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
	"k8s.io/client-go/rest"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration files",
	Long:  `Validate the configuration files`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config.SetLogger(logger)
		var aos config.ApiObjectStore
		var err error
		if len(args) == 1 {
			logger.Info("reading config files", "dir", args[0])
			aos, err = config.ReadLocalManifests(args[0], options)
			if err != nil {
				return err
			}
		} else {
			var clientConfig *rest.Config
			aos, clientConfig, err = config.ConnectToKube(options)
			if err != nil {
				return err
			}
			logger.Info("connecting to Kubernetes for resources",
				"host", clientConfig.Host)
		}
		err = config.ValidateConsistency(aos)
		if err == nil {
			logger.Info("config files are valid")
		} else {
			logger.V(-1).Info("config files are not valid",
				"error", err.Error())
		}
		return nil

	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

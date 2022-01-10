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
	"context"
	"time"

	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/operator"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

var dryRun bool
var options *cliOptions

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply the configuration",
	Long: `The necessary configuration steps are performed based
on the configuration files which describe the expected
state of the system.`,
	Args: cobra.MaximumNArgs(1),
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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		return operator.FullResync(ctx, aos, dryRun)
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	options = &cliOptions{}
	applyCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "if specified, no operation will be performed")
	applyCmd.PersistentFlags().BoolVar(&options.forceDelete, "force-delete", false, "if specified, projects will be deleted, even with repositories")
}

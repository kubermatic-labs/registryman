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
	"errors"
	"fmt"

	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry/reconciler"
	"github.com/spf13/cobra"
)

var dryRun bool
var options *cliOptions

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("apply called")

		logger.Info("reading config files", "dir", args[0])
		config.SetLogger(logger)

		manifests, err := config.ReadManifests(args[0], options)

		if err != nil {
			return err
		}

		expectedRegistries := manifests.ExpectedProvider().GetRegistries()
		sideeffectCtx := context.WithValue(context.Background(), reconciler.SideEffectManifestManipulator, manifests)
		for _, expectedRegistry := range expectedRegistries {
			logger.Info("inspecting registry", "registry_name", expectedRegistry.GetName())
			regStatusExpected, err := reconciler.GetRegistryStatus(expectedRegistry)
			if err != nil {
				return err
			}
			logger.V(1).Info("expected registry status acquired", "status", regStatusExpected)
			actualRegistry, err := expectedRegistry.ToReal(logger)
			if err != nil {
				return err
			}
			regStatusActual, err := reconciler.GetRegistryStatus(actualRegistry)
			if err != nil {
				return err
			}
			logger.V(1).Info("actual registry status acquired", "status", regStatusActual)
			actions := reconciler.Compare(manifests.ExpectedProvider(), regStatusActual, regStatusExpected)
			logger.Info("ACTIONS:")
			for _, action := range actions {
				if !dryRun {
					logger.Info(action.String())
					sideEffect, err := action.Perform(actualRegistry)
					if err != nil {
						if errors.Is(err, globalregistry.ErrRecoverableError) {
							logger.V(-1).Info(err.Error())
						} else {
							return err
						}
					}
					if err = sideEffect.Perform(sideeffectCtx); err != nil {
						return err
					}
				} else {
					logger.Info(action.String(), "dry-run", dryRun)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	options = &cliOptions{}
	applyCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "if specified, no operation will be performed")
	applyCmd.PersistentFlags().BoolVar(&options.forceDelete, "force-delete", false, "if specified, projects will be deleted, even with repositories")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// applyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// applyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

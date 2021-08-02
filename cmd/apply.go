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
	Short: "Apply the configuration",
	Long: `The necessary configuration steps are performed based
on the configuration files which describe the expected
state of the system.`,
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
}

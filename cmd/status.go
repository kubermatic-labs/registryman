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
	"encoding/json"
	"fmt"
	"os"

	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry/reconciler"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get registry status information",
	Long:  `Get registry status information`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		config.SetLogger(logger)
		var aos config.ApiObjectStore
		var err error
		if len(args) == 1 {
			logger.Info("reading config files", "dir", args[0])
			aos, err = config.ReadLocalManifests(args[0], nil)
			if err != nil {
				return err
			}
		} else {
			var clientConfig *rest.Config
			aos, clientConfig, err = config.ConnectToKube(nil)
			if err != nil {
				return err
			}
			logger.Info("connecting to Kubernetes for resources",
				"host", clientConfig.Host)
		}

		expectedRegistries := config.NewExpectedProvider(aos).GetRegistries()
		for _, expectedRegistry := range expectedRegistries {
			fmt.Println("#")
			fmt.Println("#")
			fmt.Printf("# %s\n", expectedRegistry.GetName())
			fmt.Println("#")
			fmt.Println("#")
			actualRegistry, err := expectedRegistry.ToReal()
			if err != nil {
				return err
			}
			regStatusActual, err := reconciler.GetRegistryStatus(actualRegistry)
			if err != nil {
				return err
			}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			err = encoder.Encode(regStatusActual)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

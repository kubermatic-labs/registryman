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
	"encoding/json"
	"fmt"
	"os"
	"time"

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry/reconciler"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"
)

var filteredRegistries []string
var outputEncoder string

func registryInScope(registryName string) bool {
	if len(filteredRegistries) == 0 {
		return true
	}
	for _, filteredRegistry := range filteredRegistries {
		if registryName == filteredRegistry {
			return true
		}
	}
	return false
}

type encoder interface {
	Encode(v interface{}) (err error)
}

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

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		expectedRegistries := config.NewExpectedProvider(aos).GetRegistries(ctx)
		registryStatuses := map[string]*api.RegistryStatus{}
		for _, expectedRegistry := range expectedRegistries {
			if !registryInScope(expectedRegistry.GetName()) {
				continue
			}
			actualRegistry, err := expectedRegistry.ToReal()
			if err != nil {
				return err
			}
			registryStatuses[expectedRegistry.GetName()], err = reconciler.GetRegistryStatus(ctx, actualRegistry)
			if err != nil {
				return err
			}
		}
		var enc encoder
		switch outputEncoder {
		case "json":
			enc = json.NewEncoder(os.Stdout)
			enc.(*json.Encoder).SetIndent("", "  ")
		case "yaml":
			enc = yaml.NewEncoder(os.Stdout)
		default:
			return fmt.Errorf("invalid output format: %s", outputEncoder)
		}
		err = enc.Encode(registryStatuses)

		return err
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
	statusCmd.PersistentFlags().StringSliceVarP(&filteredRegistries,
		"registries",
		"r",
		[]string{}, "Select which registries shall be checked. When not set all registries will be checked.")
	statusCmd.PersistentFlags().StringVarP(&outputEncoder, "output", "o", "json", "Output format. Supported values are json or yaml.")
}

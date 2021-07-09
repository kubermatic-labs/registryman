/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"fmt"
	"path"
	"strings"

	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/docker"
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Uploads a repository from a local directory to a registry",
	Long: `The import command takes two arguments, the path to the 
local directory that contains the repository in .tar format, and also
the URL of the registry, where the repository will be pushed.
	`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("import called")
		importPath := args[0]
		configDir := args[1]

		config.SetLogger(logger)

		splitN := strings.SplitN(path.Clean(importPath), "/", -1)
		destinationRepo := path.Join(splitN[len(splitN)-2], splitN[len(splitN)-1])

		manifests, err := config.ReadManifests(configDir, nil)
		if err != nil {
			return err
		}

		expectedRegistries := manifests.ExpectedProvider().GetRegistries()
		for _, expectedRegistry := range expectedRegistries {
			logger.Info("inspecting registry", "registry_name", expectedRegistry.GetName())

			actualRegistry, err := expectedRegistry.ToReal(logger)
			if err != nil {
				return err
			}
			transfer := docker.New(actualRegistry.GetUsername(), actualRegistry.GetPassword())

			if err := transfer.Import(importPath, destinationRepo, logger); err != nil {
				return err
			}
			logger.Info("importing project finished", "target registry", expectedRegistry.GetName())
			return nil
		}

		return fmt.Errorf("importing repository on path %s failed", importPath)
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

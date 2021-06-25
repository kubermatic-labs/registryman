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

	"github.com/kubermatic-labs/registryman/pkg/docker"
	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "It saves the given repository in tar format",
	Long: `The export command takes two arguments, the repository to be saved
	and also the path/filename of the generated tar file.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("export called")
		repository := args[0]
		destinationPath := args[1]
		// TODO: project arg instead of repository
		// TODO: default destination

		// Create the neccessary directory structure on the given storage
		// according to the projects involved

		// // TODO: adding metadata to exported tars
		// // Maybe using ImageTag()

		// // TODO: Delete image if it has been pulled

		err := docker.Export(repository, destinationPath, logger)
		if err != nil {
			return err
		}

		logger.Info("exporting finished", "result path", destinationPath)
		return nil
		// TODO: sync command
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// exportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// exportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

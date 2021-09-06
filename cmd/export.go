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
	"fmt"
	"time"

	"context"
	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	"github.com/kubermatic-labs/registryman/pkg/skopeo"
	"github.com/spf13/cobra"
)

var destinationPath string

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "It creates a backup for a given project in tar format",
	Long: `The export command takes two arguments, the name of the project to be saved
and the path for the configuration directory describing the registry. The default
path/filename of the generated tar file can also be overwritten with the '-o' flag.`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		configDir := args[1]

		logger.Info("reading config files", "dir", configDir)
		config.SetLogger(logger)

		aos, err := config.ReadLocalManifests(configDir, nil)
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		project, err := config.GetProjectByName(ctx, aos, projectName)
		if err != nil {
			return err
		}

		projectFullPath, err := project.GenerateProjectRepoName()
		if err != nil {
			return err
		}
		transfer := skopeo.New(project.Registry.GetUsername(), project.Registry.GetPassword())

		projectWithRepositories, ok := project.Project.(globalregistry.ProjectWithRepositories)
		if !ok {
			return fmt.Errorf("%s does not have repositories", projectFullPath)
		}
		repositories, err := projectWithRepositories.GetRepositories(ctx)
		if err != nil {
			return err
		}
		for _, repoName := range repositories {
			repoFullPath := fmt.Sprintf("%s/%s", projectFullPath, repoName)
			logger.Info("exporting repository", "path", repoFullPath)
			err = transfer.Export(repoFullPath, destinationPath, logger)
			if err != nil {
				return err
			}
		}
		logger.Info("exporting project finished", "result path", destinationPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.PersistentFlags().StringVarP(&destinationPath, "output", "o", "./exported-registry", "The path for the saved repositories")
}

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
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/statusupdater"
	"github.com/spf13/cobra"
)

// operatorCmd represents the operator command
var operatorCmd = &cobra.Command{
	Use:   "operator",
	Short: "Start in operator mode",
	Long:  `Start in operator mode`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("operator called")
		aos, clientConfig, err := config.ConnectToKube(options)
		if err != nil {
			logger.Error(err, "error connecting to Kubernetes API server")
			return
		}
		logger.Info("connecting to Kubernetes for resources",
			"host", clientConfig.Host)
		statusUpdater := statusupdater.New(
			logger,
			10*time.Second,
			aos,
		)
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		statusUpdater.Start(ctx)
		<-ctx.Done()
	},
}

func init() {
	rootCmd.AddCommand(operatorCmd)
}

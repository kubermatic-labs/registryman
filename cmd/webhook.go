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

	"net/http"

	"github.com/kubermatic-labs/registryman/pkg/webhook"
	"github.com/spf13/cobra"
)

var (
	webhookListenPort *int
	keyFilePath       *string
	certFilePath      *string
)

// webhookCmd represents the webhook command
var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Start validating webhook server",
	Long:  "Start validating webhook server",
	Run: func(cmd *cobra.Command, args []string) {
		webhook.SetLogger(logger)
		logger.V(1).Info("startup configuration",
			"verbose", verbose)
		http.HandleFunc("/", webhook.AdmissionRequestHandler)
		logger.Info("starting validating webhook server",
			"port", *webhookListenPort,
		)
		panic(http.ListenAndServeTLS(fmt.Sprintf(":%d", *webhookListenPort),
			*certFilePath, *keyFilePath, nil))
	},
}

func init() {
	rootCmd.AddCommand(webhookCmd)

	webhookListenPort = webhookCmd.PersistentFlags().IntP("port", "p", 443, "Port where the webhook service listens on.")
	keyFilePath = webhookCmd.Flags().StringP("key", "k", "tls/tls.key", "TLS key file path.")
	certFilePath = webhookCmd.Flags().StringP("cert", "c", "tls/tls.crt", "TLS cert file path.")
}

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

	"github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/spf13/cobra"
	"k8s.io/kube-openapi/pkg/builder"
	"k8s.io/kube-openapi/pkg/common"
	"k8s.io/kube-openapi/pkg/util"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// swaggerCmd represents the swagger command
var swaggerCmd = &cobra.Command{
	Use:   "swagger",
	Short: "Generate swagger schema",
	Long:  `Generate a swagger schema of the resources used in the config files. The generated schema is in JSON format and is sent to the standard output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		config := &common.Config{
			ProtocolList: []string{"http", "https"},
			Info: &spec.Info{
				VendorExtensible: spec.VendorExtensible{
					Extensions: map[string]interface{}{},
				},
				InfoProps: spec.InfoProps{
					Description:    "Registry API description for Registryman project",
					Title:          "RegistryMan API",
					TermsOfService: "",
					Contact: &spec.ContactInfo{
						Name: "Registryman",
						URL:  "https://github.com/kubermatic-labs/registryman",
					},
					License: &spec.License{
						Name: "Apache V2.0",
						URL:  "https://www.apache.org/licenses/LICENSE-2.0",
					},
					Version: "v1alpha1",
				},
			},
			// DefaultResponse:     &spec.Response{},
			// ResponseDefinitions: map[string]spec.Response{},
			// CommonResponses:     map[int]spec.Response{},
			// IgnorePrefixes:      []string{},
			GetDefinitions: v1alpha1.GetOpenAPIDefinitions,
			// GetOperationIDAndTags: func(r *restful.Route) (string, []string, error) {
			// },
			// GetDefinitionName: func(name string) (string, spec.Extensions) {
			// },
			// PostProcessSpec: func(*spec.Swagger) (*spec.Swagger, error) {
			// },
			// SecurityDefinitions: &map[string]*spec.SecurityScheme{},
			// DefaultSecurity:     []map[string][]string{},
		}
		swagger, err := builder.BuildOpenAPIDefinitionsForResources(config,
			util.GetCanonicalTypeName(&v1alpha1.Registry{}),
			util.GetCanonicalTypeName(&v1alpha1.Project{}),
		)
		if err != nil {
			return err
		}
		b, err := swagger.MarshalJSON()
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(swaggerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// swaggerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// swaggerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

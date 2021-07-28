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

package config

import (
	"context"

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	regmanclient "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/clientset/versioned"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	"github.com/spf13/pflag"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var clientConfig *rest.Config
var kubeConfig clientcmd.ClientConfig

func init() {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	// if you want to change the loading rules (which files in which order), you can do so here

	configOverrides := &clientcmd.ConfigOverrides{}
	clientcmd.BindOverrideFlags(configOverrides, pflag.CommandLine,
		clientcmd.RecommendedConfigOverrideFlags(""))
	// if you want to change override values or bind them to flags, there are methods to help you

	kubeConfig = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
}

type kubeApiObjectStore struct {
	regmanClient *regmanclient.Clientset
}

var _ ApiObjectStore = &kubeApiObjectStore{}

func ConnectToKube() (ApiObjectStore, *rest.Config, error) {
	var err error
	clientConfig, err = kubeConfig.ClientConfig()
	if err != nil {
		return nil, nil, err
	}
	return &kubeApiObjectStore{
		regmanClient: regmanclient.NewForConfigOrDie(clientConfig),
	}, clientConfig, nil
}

// WriteResource serializes the object specified by the obj parameter.
// The filename parameter specifies the name of the file to be created.
// The path where the file is created is set when the ReadManifests
// function creates the ApiObjectStore.
func (aos *kubeApiObjectStore) WriteResource(obj runtime.Object) error {
	panic("not implemented") // TODO: Implement
}

// RemoveResource removes the file from the filesystem. The path where
// the file is removed from is set when the ReadManifests function
// creates the ApiObjectStore.
func (aos *kubeApiObjectStore) RemoveResource(objectName string) error {
	panic("not implemented") // TODO: Implement
}

// GetRegistries returns the parsed registries as API objects.
func (aos *kubeApiObjectStore) GetRegistries() []*api.Registry {
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		panic(err)
	}
	registryList, err := aos.regmanClient.RegistrymanV1alpha1().Registries(namespace).List(context.Background(), v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	apiRegistries := make([]*api.Registry, len(registryList.Items))
	for i, reg := range registryList.Items {
		apiRegistries[i] = &reg
	}
	return apiRegistries
}

// GetProjects returns the parsed projects as API objects.
func (aos *kubeApiObjectStore) GetProjects() []*api.Project {
	panic("not implemented") // TODO: Implement
}

// GetScanners returns the parsed scanners as API objects.
func (aos *kubeApiObjectStore) GetScanners() []*api.Scanner {
	panic("not implemented") // TODO: Implement
}

// GetGlobalRegistryOptions returns the ApiObjectStore related CLI options of an
// apply.
func (aos *kubeApiObjectStore) GetGlobalRegistryOptions() globalregistry.RegistryOptions {
	panic("not implemented") // TODO: Implement
}

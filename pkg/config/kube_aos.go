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
	"fmt"

	"github.com/go-logr/logr"
	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	regmanclient "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/clientset/versioned"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	applyCoreV1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var clientConfig *rest.Config
var kubeConfig clientcmd.ClientConfig

const fieldManager = "regman"

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
	kubeClient   *kubernetes.Clientset
	options      globalregistry.RegistryOptions
}

var _ ApiObjectStore = &kubeApiObjectStore{}

func ConnectToKube(options globalregistry.RegistryOptions) (ApiObjectStore, *rest.Config, error) {
	var err error
	clientConfig, err = kubeConfig.ClientConfig()
	if err != nil {
		return nil, nil, err
	}
	return &kubeApiObjectStore{
		options:      options,
		regmanClient: regmanclient.NewForConfigOrDie(clientConfig),
		kubeClient:   kubernetes.NewForConfigOrDie(clientConfig),
	}, clientConfig, nil
}

// WriteResource serializes the object specified by the obj parameter.
// The filename parameter specifies the name of the file to be created.
// The path where the file is created is set when the ReadManifests
// function creates the ApiObjectStore.
func (aos *kubeApiObjectStore) WriteResource(ctx context.Context, obj runtime.Object) error {
	gvk := obj.GetObjectKind().GroupVersionKind()
	logger.V(1).Info("WriteResource",
		"group", gvk.Group,
		"version", gvk.Version,
		"kind", gvk.Kind,
	)
	switch gvk {
	default:
		logger.V(-1).Info("WriterResource invoked, unsupported resource",
			"group", gvk.Group,
			"version", gvk.Version,
			"kind", gvk.Kind,
		)
	case schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Secret",
	}:
		secret := obj.(*corev1.Secret)
		namespace, _, err := kubeConfig.Namespace()
		if err != nil {
			return fmt.Errorf("cannot get Kubernetes namespace: %w", err)
		}
		logger.V(1).Info("creating a new secret",
			"name", secret.GetName(),
		)
		applyConfig := applyCoreV1.Secret(secret.Name, namespace).
			WithData(secret.Data).
			WithStringData(secret.StringData).
			WithType(secret.Type)
		if err != nil {
			return fmt.Errorf("error creating SecretApplyConfiguration: %w", err)
		}
		_, err = aos.kubeClient.CoreV1().Secrets(namespace).Apply(ctx,
			applyConfig,
			v1.ApplyOptions{
				FieldManager: fieldManager,
			})
		if err != nil {
			return fmt.Errorf("error applying secret: %w", err)
		}
	}
	return nil
}

// RemoveResource removes the file from the filesystem. The path where
// the file is removed from is set when the ReadManifests function
// creates the ApiObjectStore.
func (aos *kubeApiObjectStore) RemoveResource(ctx context.Context, obj runtime.Object) error {
	gvk := obj.GetObjectKind().GroupVersionKind()
	logger.V(1).Info("WriteResource",
		"group", gvk.Group,
		"version", gvk.Version,
		"kind", gvk.Kind,
	)
	switch gvk {
	default:
		logger.V(-1).Info("WriterResource invoked, unsupported resource",
			"group", gvk.Group,
			"version", gvk.Version,
			"kind", gvk.Kind,
		)
	case schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Secret",
	}:
		secret := obj.(*corev1.Secret)
		namespace, _, err := kubeConfig.Namespace()
		if err != nil {
			return fmt.Errorf("cannot get Kubernetes namespace: %w", err)
		}
		logger.V(1).Info("removing secret",
			"name", secret.GetName(),
		)
		err = aos.kubeClient.CoreV1().Secrets(namespace).Delete(ctx, secret.GetName(), v1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("error removing secret: %w", err)
		}
	}
	return nil
}

// GetRegistries returns the parsed registries as API objects.
func (aos *kubeApiObjectStore) GetRegistries(ctx context.Context) []*api.Registry {
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		panic(err)
	}
	registryList, err := aos.regmanClient.RegistrymanV1alpha1().Registries(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	apiRegistries := make([]*api.Registry, len(registryList.Items))
	for i := range registryList.Items {
		apiRegistries[i] = &registryList.Items[i]
	}
	return apiRegistries
}

// GetProjects returns the parsed projects as API objects.
func (aos *kubeApiObjectStore) GetProjects(ctx context.Context) []*api.Project {
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		panic(err)
	}
	projectList, err := aos.regmanClient.RegistrymanV1alpha1().Projects(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	apiProjects := make([]*api.Project, len(projectList.Items))
	for i := range projectList.Items {
		apiProjects[i] = &projectList.Items[i]
	}
	return apiProjects
}

// GetScanners returns the parsed scanners as API objects.
func (aos *kubeApiObjectStore) GetScanners(ctx context.Context) []*api.Scanner {
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		panic(err)
	}
	scannerList, err := aos.regmanClient.RegistrymanV1alpha1().Scanners(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		panic(err)
	}
	apiScanners := make([]*api.Scanner, len(scannerList.Items))
	for i := range scannerList.Items {
		apiScanners[i] = &scannerList.Items[i]
	}
	return apiScanners
}

// GetGlobalRegistryOptions returns the ApiObjectStore related CLI options of an
// apply.
func (aos *kubeApiObjectStore) GetGlobalRegistryOptions() globalregistry.RegistryOptions {
	return aos.options
}

func (aos *kubeApiObjectStore) GetLogger() logr.Logger {
	return logger
}

func (aos *kubeApiObjectStore) UpdateRegistryStatus(ctx context.Context, reg *api.Registry) error {
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		return err
	}
	_, err = aos.regmanClient.RegistrymanV1alpha1().Registries(namespace).UpdateStatus(ctx, reg, v1.UpdateOptions{
		FieldManager: fieldManager,
	})
	return err
}

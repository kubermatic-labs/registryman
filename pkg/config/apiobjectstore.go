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

	"github.com/go-logr/logr"
	_ "github.com/kubermatic-labs/registryman/pkg/acr"
	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	_ "github.com/kubermatic-labs/registryman/pkg/harbor"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
	}
	err := api.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	scheme.AddKnownTypeWithName(secret.GroupVersionKind(), secret)
}

// ApiObjectStore interface is an abstract interface that hides the difference
// between the local file and Kubernetes resource based config management.
type ApiObjectStore interface {
	// WriteResource serializes the object specified by the obj parameter.
	WriteResource(ctx context.Context, obj runtime.Object) error

	// RemoveResource removes the file from the filesystem.
	RemoveResource(ctx context.Context, obj runtime.Object) error

	// GetRegistries returns the parsed registries as API objects.
	GetRegistries(context.Context) []*api.Registry

	// GetProjects returns the parsed projects as API objects.
	GetProjects(context.Context) []*api.Project

	// GetScanners returns the parsed scanners as API objects.
	GetScanners(context.Context) []*api.Scanner

	// GetGlobalRegistryOptions returns the ApiObjectStore related CLI options of an
	// apply.
	GetGlobalRegistryOptions() globalregistry.RegistryOptions

	// GetLogger returns the logr.Logger interface that the ApiObjectStore is using
	// for logging.
	GetLogger() logr.Logger

	// UpdateRegistryStatus persists the registry status of the given
	// Registry resource.
	UpdateRegistryStatus(context.Context, *api.Registry) error
}

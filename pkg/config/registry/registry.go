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

// package registry implements the globalregistry.Registry interface based on
// the registryman.kubermatic.com/v1alpha1 API objects.
package registry

import (
	"context"
	"strconv"

	"github.com/go-logr/logr"
	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

// ApiObjectProvider interface describes the methods that are needed to get the
// different API resources.
type ApiObjectProvider interface {
	GetProjects(context.Context) []*api.Project
	GetRegistries(context.Context) []*api.Registry
	GetScanners(context.Context) []*api.Scanner
	GetGlobalRegistryOptions() globalregistry.RegistryOptions
	GetLogger() logr.Logger
}

// Registry type describes the API representation of a registry (i.e. the
// expected state) but it also implements the globalregistry.Registry interface.
type Registry struct {
	apiProvider ApiObjectProvider
	apiRegistry *api.Registry
}

var _ globalregistry.Registry = &Registry{}

// New function creates a new Registry value from the API representation of the
// registry.
func New(reg *api.Registry, store ApiObjectProvider) *Registry {
	return &Registry{
		apiProvider: store,
		apiRegistry: reg,
	}
}

// GetName method implements the globalregistry.RegistryConfig interface.
func (reg *Registry) GetName() string {
	return reg.apiRegistry.GetName()
}

// GetProvider method implements the globalregistry.RegistryConfig interface.
func (reg *Registry) GetProvider() string {
	return reg.apiRegistry.Spec.Provider
}

// GetAPIEndpoint method implements the globalregistry.RegistryConfig interface.
func (reg *Registry) GetAPIEndpoint() string {
	return reg.apiRegistry.Spec.APIEndpoint
}

// GetUsername method implements the globalregistry.RegistryConfig interface.
func (reg *Registry) GetUsername() string {
	return reg.apiRegistry.Spec.Username
}

// GetPassword method implements the globalregistry.RegistryConfig interface.
func (reg *Registry) GetPassword() string {
	return reg.apiRegistry.Spec.Password
}

type registryOptions struct {
	forceDelete bool
}

var _ globalregistry.CanForceDelete = &registryOptions{}

// ForceDeleteProjects returns with the value of the force-delete option.
func (o *registryOptions) ForceDeleteProjects() bool {
	return o.forceDelete
}

// GetOptions method implements the globalregistry.RegistryConfig interface. The
// method returns the RegistryOptions configured via annotations of the Registry
// object. If there are no annotations, the CLI options of the API provider is
// used.
//
// Supported annotations:
// - registryman.kubermatic.com/forceDelete: <bool_as_string>
func (reg *Registry) GetOptions() globalregistry.RegistryOptions {
	if val, ok := reg.apiRegistry.Annotations["registryman.kubermatic.com/forceDelete"]; ok {
		b, err := strconv.ParseBool(val)
		if err != nil {
			reg.apiProvider.GetLogger().V(-1).Info("invalid value for registryman.kubermatic.com/forceDelete annotation, expected \"true\" or \"false\"",
				"registry", reg.apiRegistry.GetName(),
				"value", val)
			return &registryOptions{forceDelete: false}
		}
		return &registryOptions{forceDelete: b}
	}
	return reg.apiProvider.GetGlobalRegistryOptions()
}

// ToReal method turns the (i.e. expected) Registry value into a
// provider-specific (i.e. actual) registry value.
func (reg *Registry) ToReal() (globalregistry.Registry, error) {
	return globalregistry.New(reg.apiProvider.GetLogger(), reg)
}

func (reg *Registry) registryCapabilities() registryCapabilities {
	return registryCapabilities{
		isGlobal:                reg.apiRegistry.Spec.Role == "GlobalHub",
		ReplicationCapabilities: globalregistry.GetReplicationCapability(reg.GetProvider()),
	}
}

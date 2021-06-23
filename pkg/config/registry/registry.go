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

package registry

import (
	"github.com/go-logr/logr"
	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman.kubermatic.com/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

// ApiObjectProvider interface describes the methods that are needed to get the
// different API resources.
type ApiObjectProvider interface {
	GetProjects() []*api.Project
	GetRegistries() []*api.Registry
	GetScanners() []*api.Scanner
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

// ReplicationAPI method implements the globalregistry.Registry interface.
func (reg *Registry) ReplicationAPI() globalregistry.ReplicationAPI {
	return &replicationAPI{}
}

// ProjectAPI method implements the globalregistry.Registry interface.
func (reg *Registry) ProjectAPI() globalregistry.ProjectAPI {
	return &projectAPI{
		registry: reg,
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

// ToReal method turns the (i.e. expected) Registry value into a
// provider-specific (i.e. actual) registry value.
func (reg *Registry) ToReal(logger logr.Logger) (globalregistry.Registry, error) {
	return globalregistry.New(logger, reg)
}

func (reg *Registry) registryCapabilities() registryCapabilities {
	return registryCapabilities{
		isGlobal:                reg.apiRegistry.Spec.Role == "GlobalHub",
		ReplicationCapabilities: globalregistry.GetReplicationCapability(reg.GetProvider()),
	}
}

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
	api "github.com/kubermatic-labs/registryman/pkg/apis/globalregistry/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type ApiProvider interface {
	GetProjects() []*api.Project
	GetRegistries() []*api.Registry
}

type Registry struct {
	ApiProvider
	*api.Registry
}

var _ globalregistry.Registry = &Registry{}

func New(reg *api.Registry, store ApiProvider) *Registry {
	return &Registry{
		store,
		reg,
	}
}

func (reg *Registry) ReplicationAPI() globalregistry.ReplicationAPI {
	return &replicationAPI{}
}

func (reg *Registry) ProjectAPI() globalregistry.ProjectAPI {
	return &projectAPI{
		registry: reg,
	}
}

func (reg *Registry) GetProvider() string {
	return reg.Spec.Provider
}

func (reg *Registry) GetAPIEndpoint() string {
	return reg.Spec.APIEndpoint
}

func (reg *Registry) GetUsername() string {
	return reg.Spec.Username
}

func (reg *Registry) GetPassword() string {
	return reg.Spec.Password
}

func (reg *Registry) ToReal(logger logr.Logger) (globalregistry.Registry, error) {
	return globalregistry.New(logger, reg)
}

func (reg *Registry) registryCapabilities() registryCapabilities {
	return registryCapabilities{
		isGlobal:                reg.Spec.Role == "GlobalHub",
		ReplicationCapabilities: globalregistry.GetReplicationCapability(reg.GetProvider()),
	}
}

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

	"github.com/kubermatic-labs/registryman/pkg/config/registry"
)

// ExpectedProvider is a database of the resources which implement the
// interfaces defines in the globalregistry package.
//
// The resources in the database usually show the expected state of the
// resources.
type ExpectedProvider struct {
	registry.ApiObjectProvider
}

// NewExpectedProvider method turns an ApiObjectStore into an ExpectedProvider.
func NewExpectedProvider(aop registry.ApiObjectProvider) *ExpectedProvider {
	return &ExpectedProvider{aop}
}

// GetRegistries returns the Registries of the resource database.
func (expp ExpectedProvider) GetRegistries(ctx context.Context) []*registry.Registry {
	apiRegistries := expp.ApiObjectProvider.GetRegistries(ctx)
	registries := make([]*registry.Registry, len(apiRegistries))
	for i, apiRegistry := range apiRegistries {
		registries[i] = registry.New(apiRegistry, expp.ApiObjectProvider)
	}
	return registries
}

// GetRegistryByName returns a Registry with the given name from the database.
// If no Registry if found with the specified name, nil is returned.
func (expp *ExpectedProvider) GetRegistryByName(ctx context.Context, name string) *registry.Registry {
	for _, apiRegistry := range expp.ApiObjectProvider.GetRegistries(ctx) {
		if apiRegistry.GetName() == name {
			return registry.New(apiRegistry, expp.ApiObjectProvider)
		}
	}
	return nil
}

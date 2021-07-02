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

package globalregistry

import (
	"fmt"

	"github.com/go-logr/logr"
)

// RegistryCreator function type can be used to create a Registry interface.
type RegistryCreator func(logr.Logger, RegistryConfig) (Registry, error)

var (
	registeredRegistryCreators        map[string]RegistryCreator
	registeredReplicationCapabilities map[string]ReplicationCapabilities
)

func init() {
	registeredRegistryCreators = make(map[string]RegistryCreator)
	registeredReplicationCapabilities = make(map[string]ReplicationCapabilities)
}

// ReplicationCapabilities interface defines the methods that show the
// replication capabilities of a registry provider.
type ReplicationCapabilities interface {
	CanPull() bool
	CanPush() bool
}

// GetReplicationCapability function returns the ReplicationCapabilities of a
// registered registry provider.
func GetReplicationCapability(provider string) ReplicationCapabilities {
	cap, found := registeredReplicationCapabilities[provider]
	if !found {
		panic(fmt.Sprintf("provider %s is not registered", provider))
	}
	return cap
}

// RegistryConfig interface describes a registry configuration that is needed to
// create a new provider-specific Registry via its constructor.
type RegistryConfig interface {
	GetProvider() string
	GetUsername() string
	GetPassword() string
	GetAPIEndpoint() string
	GetName() string
	GetOptions() RegistryOptions
}

// Registry is an abstraction over registries. It is an abstraction over the
// real, provider-specific registries and over the expected registry states
// defined by the api.Registry type.
type Registry interface {
	RegistryConfig
	ReplicationAPI() ReplicationAPI
	ProjectAPI() ProjectAPI
}

// New creates a provider specific Registry. The provider must be registered
// first. If the provider is not registered, an error is returned. Otherwise the
// constructor function of the registered provider is invoked.
func New(logger logr.Logger, config RegistryConfig) (Registry, error) {
	registryProvider := config.GetProvider()
	constructor, ok := registeredRegistryCreators[registryProvider]
	if !ok {
		return nil, fmt.Errorf("registry type %s not known", registryProvider)
	}
	return constructor(logger, config)
}

// RegisterProviderImplementation is used by the different Registry interface
// implementations to register the Register constructors. After a constructor
// function is registered, a new Registry can be created using the New function.
func RegisterProviderImplementation(providerName string,
	constructor RegistryCreator,
	repCap ReplicationCapabilities,
) {
	registeredRegistryCreators[providerName] = constructor
	registeredReplicationCapabilities[providerName] = repCap
}

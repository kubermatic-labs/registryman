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

// acr package implements the globalregistry.Registry interface for the registry
// provider Azure Container Registry.
package acr

import (
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

const path = "/v2/_catalog"

type registry struct {
	logger    logr.Logger
	parsedUrl *url.URL
	globalregistry.Registry
	*http.Client
}

type repositories struct {
	Repositories []string `json:"repositories,omitempty"`
}

var _ globalregistry.Registry = &registry{}

func init() {
	globalregistry.RegisterProviderImplementation(
		"acr",
		newRegistry,
		acrRegistryCapabilities{},
	)
}

func newRegistry(logger logr.Logger, config globalregistry.Registry) (globalregistry.Registry, error) {
	var err error
	r := &registry{
		Registry: config,
		Client:   http.DefaultClient,
		logger:   logger,
	}

	r.parsedUrl, err = url.Parse(config.GetAPIEndpoint())
	if err != nil {
		return nil, err
	}
	return r, nil
}

type acrRegistryCapabilities struct{}

var _ globalregistry.ReplicationCapabilities = acrRegistryCapabilities{}

func (cap acrRegistryCapabilities) CanPull() bool {
	return false
}

func (cap acrRegistryCapabilities) CanPush() bool {
	return false
}

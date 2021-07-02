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
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

const path = "/v2/_catalog"

type registry struct {
	logger logr.Logger
	globalregistry.RegistryConfig
	projects     *projectAPI
	replications *replicationAPI
	*http.Client
	parsedUrl *url.URL
}

type replicationAPI struct{}

type acrRegistryCapabilities struct{}

type repositories struct {
	Repositories []string `json:"repositories,omitempty"`
}

var _ globalregistry.Registry = &registry{}
var _ globalregistry.ReplicationCapabilities = acrRegistryCapabilities{}

func init() {
	globalregistry.RegisterProviderImplementation(
		"acr",
		newRegistry,
		acrRegistryCapabilities{},
	)
}

func newRegistry(logger logr.Logger, config globalregistry.RegistryConfig) (globalregistry.Registry, error) {
	var err error
	r := &registry{
		RegistryConfig: config,
		projects:       &projectAPI{},
		replications:   &replicationAPI{},
		Client:         http.DefaultClient,
		logger:         logger,
	}
	r.projects, err = newProjectAPI(r)
	if err != nil {
		return nil, err
	}

	r.parsedUrl, err = url.Parse(config.GetAPIEndpoint())
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *registry) ProjectAPI() globalregistry.ProjectAPI {
	return r.projects
}

func (r *registry) ReplicationAPI() globalregistry.ReplicationAPI {
	return r.replications
}

func (r *replicationAPI) List() ([]globalregistry.ReplicationRule, error) {
	return nil, fmt.Errorf("replicationAPI.List not implemented")
}

func (cap acrRegistryCapabilities) CanPull() bool {
	return false
}

func (cap acrRegistryCapabilities) CanPush() bool {
	return false
}

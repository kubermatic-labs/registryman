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

// artifactory package implements the globalregistry.Registry interface for the registry
// provider Azure Container Registry.
package artifactory

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"path"

	"github.com/go-logr/logr"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

const catalogPath = "/v2/_catalog"

type registry struct {
	logger       logr.Logger
	parsedUrl    *url.URL
	registryPath string
	name         string
	globalregistry.Registry
	*http.Client
}

type repositories struct {
	Repositories []string `json:"repositories,omitempty"`
}

var _ globalregistry.Registry = &registry{}

func init() {
	globalregistry.RegisterProviderImplementation(
		"artifactory",
		newRegistry,
		artifactoryRegistryCapabilities{},
	)
}

func newRegistry(logger logr.Logger, config globalregistry.Registry) (globalregistry.Registry, error) {
	// TODO: delete this hack
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: customTransport}
	var err error
	r := &registry{
		Registry: config,
		Client:   client,
		logger:   logger,
	}

	r.parsedUrl, err = url.Parse(config.GetAPIEndpoint())
	r.name = path.Base(r.parsedUrl.Path)
	r.registryPath = r.parsedUrl.Path
	if err != nil {
		return nil, err
	}
	return r, nil
}

type artifactoryRegistryCapabilities struct{}

var _ globalregistry.ReplicationCapabilities = artifactoryRegistryCapabilities{}

func (cap artifactoryRegistryCapabilities) CanPull() bool {
	return false
}

func (cap artifactoryRegistryCapabilities) CanPush() bool {
	return false
}

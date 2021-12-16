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
// provider JFrog Artifactory.
package artifactory

import (
	"crypto/tls"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/kubermatic-labs/registryman/pkg/artifactory/pathbased"
	"github.com/kubermatic-labs/registryman/pkg/artifactory/projectbased"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

func init() {
	// during init the artifactory provider is registered
	globalregistry.RegisterProviderImplementation(
		"artifactory",
		newRegistry,
		artifactoryRegistryCapabilities{},
	)
}

// newRegistry is the constructor if the registry type. It is a globalregistry RegistryCreator.
func newRegistry(logger logr.Logger, config globalregistry.Registry) (globalregistry.Registry, error) {
	var err error
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	client := &http.Client{Transport: customTransport}

	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		for key, val := range via[0].Header {
			req.Header[key] = val
		}
		return err
	}

	dockerRegistryName := ""
	if val, ok := config.GetAnnotations()["registryman.kubermatic.com/dockerRegistryName"]; ok {
		dockerRegistryName = val
	}

	if dockerRegistryName == "" {
		accessToken := ""
		if val, ok := config.GetAnnotations()["registryman.kubermatic.com/accessToken"]; ok {
			accessToken = val

			c, err := projectbased.NewRegistry(
				logger,
				client,
				config,
				accessToken)
			if err != nil {
				return nil, err
			}
			return c, nil
		}
	}
	c, err := pathbased.NewRegistry(
		logger,
		client,
		config,
		dockerRegistryName)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type artifactoryRegistryCapabilities struct{}

func (cap artifactoryRegistryCapabilities) CanPull() bool {
	return false
}

func (cap artifactoryRegistryCapabilities) CanPush() bool {
	return false
}

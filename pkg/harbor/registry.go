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
package harbor

import (
	"bytes"
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

func init() {
	// during init the harbor provider is registered
	globalregistry.RegisterProviderImplementation(
		"harbor",
		newRegistry,
		harborRegistryCapabilities{},
	)
}

type registry struct {
	logger logr.Logger
	globalregistry.RegistryConfig
	*http.Client
	projects         *projectAPI
	remoteRegistries *remoteRegistries
	replications     *replicationAPI
	parsedUrl        *url.URL
	scanners         *scannerAPI
}

// registry type implements the globalregistry.Registry interface
var _ globalregistry.Registry = &registry{}

// newRegistry is the constructor if the registry type. It is a globalregistry RegistryCreator.
func newRegistry(logger logr.Logger, config globalregistry.RegistryConfig) (globalregistry.Registry, error) {
	var err error
	c := &registry{
		logger:         logger,
		RegistryConfig: config,
		Client:         http.DefaultClient,
	}
	c.projects, err = newProjectAPI(c)
	if err != nil {
		return nil, err
	}
	c.remoteRegistries = newRemoteRegistries(c)
	c.replications = newReplicationAPI(c)
	c.parsedUrl, err = url.Parse(config.GetAPIEndpoint())
	c.scanners, err = newScannerAPI(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *registry) ProjectAPI() globalregistry.ProjectAPI {
	return r.projects
}

func (r *registry) ReplicationAPI() globalregistry.ReplicationAPI {
	return r.replications
}

func (r *registry) ScannerAPI() globalregistry.ScannerAPI {
	return r.scanners
}

type bytesBody struct {
	*bytes.Buffer
}

func (bb bytesBody) Close() error { return nil }

// do method of Registry will perform a normal http.Registry do operation plus
// it prints extra information in case of unexpected response codes. The
// response body is replaced with a bytesBody which provides the bytes.Buffer
// (e.g. String()) methods too.
func (s *registry) do(req *http.Request) (*http.Response, error) {
	resp, err := s.Client.Do(req)
	if err != nil {
		s.logger.Error(err, "http.Client cannot Do",
			"req-url", req.URL,
		)
		return nil, err
	}

	buf := bytesBody{
		Buffer: new(bytes.Buffer),
	}
	n, err := buf.ReadFrom(resp.Body)
	if err != nil {
		s.logger.Error(err, "cannot read HTTP response body")
		return nil, err
	}
	resp.Body = buf

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		s.logger.V(-1).Info("HTTP response status code is not OK",
			"status-code", resp.StatusCode,
			"resp-body-size", n,
			"req-url", req.URL,
		)
		s.logger.V(1).Info(buf.String())
	}
	return resp, nil
}

type harborRegistryCapabilities struct{}

func (cap harborRegistryCapabilities) CanPull() bool {
	return true
}

func (cap harborRegistryCapabilities) CanPush() bool {
	return true
}

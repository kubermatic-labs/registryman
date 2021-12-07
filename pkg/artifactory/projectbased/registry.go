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

// projectbased package implements the globalregistry.Registry interface for the registry
// provider JFrog Artifactory, using the Project functionality of Artifactory 7.x.
package projectbased

import (
	"bytes"
	"context"
	"net/http"
	"net/url"

	"github.com/go-logr/logr"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type projectRegistry struct {
	logger    logr.Logger
	parsedUrl *url.URL
	globalregistry.Registry
	*http.Client

	// accessToken is the manually created token in Artifactory
	accessToken string
}

var _ globalregistry.Registry = &projectRegistry{}
var _ globalregistry.RegistryWithProjects = &projectRegistry{}

func (reg *projectRegistry) getAccessToken() string {
	return reg.accessToken
}

// newRegistry is the constructor if the registry type. It is a globalregistry RegistryCreator.
func NewRegistry(logger logr.Logger, client *http.Client, config globalregistry.Registry, token string) (globalregistry.Registry, error) {
	var err error

	c := &projectRegistry{
		logger:      logger,
		Registry:    config,
		Client:      client,
		accessToken: token}
	c.parsedUrl, err = url.Parse(config.GetAPIEndpoint())
	if err != nil {
		return nil, err
	}

	return c, nil

}

type bytesBody struct {
	*bytes.Buffer
}

func (bb bytesBody) Close() error { return nil }

// do method of Registry will perform a normal http.Registry do operation plus
// it prints extra information in case of unexpected response codes. The
// response body is replaced with a bytesBody which provides the bytes.Buffer
// (e.g. String()) methods too.
func (r *projectRegistry) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	resp, err := r.Client.Do(req)
	if err != nil {
		r.logger.Error(err, "http.Client cannot Do",
			"req-url", req.URL,
		)
		return nil, err
	}

	buf := bytesBody{
		Buffer: new(bytes.Buffer),
	}
	n, err := buf.ReadFrom(resp.Body)
	if err != nil {
		r.logger.Error(err, "cannot read HTTP response body")
		return nil, err
	}
	resp.Body = buf

	switch {
	case resp.StatusCode == 401:
		// Unauthorized
		//
		// Harbor sometimes misses to return 401 status code. It tends
		// to respond 200 even when the credentials are incorrect.
		return nil, globalregistry.ErrUnauthorized
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		// Any other error code
		r.logger.V(-1).Info("HTTP response status code is not OK",
			"status-code", resp.StatusCode,
			"resp-body-size", n,
			"req-url", req.URL,
		)
		r.logger.V(1).Info(buf.String())
	}
	return resp, nil
}

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

package pathbased

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type permissionConfiguration struct {
	Name            string     `json:"name"`
	IncludesPattern string     `json:"includesPattern"`
	ExcludesPattern string     `json:"excludesPattern"`
	Repositories    []string   `json:"repositories"`
	Principals      principals `json:"principals"`
}

type principals struct {
	Users  map[string][]string `json:"users"`
	Groups map[string][]string `json:"groups"`
}

func (r *pathRegistry) getPermission(ctx context.Context, permissionName string) (*permissionConfiguration, error) {
	apiUrl := *r.parsedUrl
	apiUrl.Path = permissionPath + "/" + permissionName
	req, err := http.NewRequest(http.MethodGet, apiUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(r.GetUsername(), r.GetPassword())
	req.Header.Add("Accept", "application/json")

	resp, err := r.do(ctx, req)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("principal target %s not exists, %w", permissionName, globalregistry.ErrRecoverableError)
		}
		return nil, err
	}

	defer resp.Body.Close()

	permission := &permissionConfiguration{}

	err = json.NewDecoder(resp.Body).Decode(&permission)
	if err != nil {
		buf := resp.Body.(*bytesBody)
		r.logger.Error(err, "json decoding failed")
		r.logger.Info(buf.String())
	}

	return permission, err
}

func (r *pathRegistry) deletePermission(ctx context.Context, projName string) error {
	permissionName := r.GetDockerRegistryName() + "_" + projName
	apiUrl := *r.parsedUrl
	apiUrl.Path = permissionPath + "/" + permissionName
	req, err := http.NewRequest(http.MethodDelete, apiUrl.String(), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(ctx, req)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("principal target %s not exists, %w", permissionName, globalregistry.ErrRecoverableError)
		}
		return err
	}

	defer resp.Body.Close()

	return nil
}

func (r *pathRegistry) createPermission(ctx context.Context, projName string, permission *permissionConfiguration) error {
	r.logger.V(1).Info("create permission target for repository",
		"ProjectName", projName,
	)

	apiUrl := *r.parsedUrl
	apiUrl.Path = permissionPath + "/" + r.GetDockerRegistryName() + "_" + projName

	reqBodyBuf := bytes.NewBuffer(nil)
	err := json.NewEncoder(reqBodyBuf).Encode(permission)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, apiUrl.String(), reqBodyBuf)
	if err != nil {
		return err
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(ctx, req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

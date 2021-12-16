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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

const registriesPath = "/api/v2.0/registries"

type registryCredential struct {
	AccessKey    string `json:"access_key"`
	AccessSecret string `json:"access_secret"`
	Type         string `json:"type"`
}

type remoteRegistryStatus struct {
	Status       string             `json:"status"`
	Credential   registryCredential `json:"credential"`
	Update_time  string             `json:"update_time"`
	Name         string             `json:"name"`
	Url          string             `url:"url"`
	Insecure     bool               `json:"insecure"`
	CreationTime string             `json:"creation_time"`
	Type         string             `json:"type"`
	Id           int                `json:"id"`
	Description  string             `json:"description"`
}

func remoteRegistryStatusFromRegistry(reg globalregistry.Registry) *remoteRegistryStatus {
	var regType string
	insecure := false
	switch reg.GetProvider() {
	case "harbor":
		regType = "harbor"
	case "acr":
		regType = "azure-acr"
	case "artifactory":
		regType = "jfrog-artifactory"
		insecure = true
	default:
		panic(fmt.Sprintf("provider %s not implemented", reg.GetProvider()))
	}
	return &remoteRegistryStatus{
		CreationTime: time.Time{}.Format(time.RFC3339),
		Update_time:  time.Time{}.Format(time.RFC3339),
		Name:         reg.GetName(),
		Url:          reg.GetAPIEndpoint(),
		Status:       "",
		Credential: registryCredential{
			AccessKey:    reg.GetUsername(),
			AccessSecret: reg.GetPassword(),
		},
		Insecure: insecure,
		Type:     regType,
		Description: fmt.Sprintf("%s is a remote %s registry",
			reg.GetName(),
			reg.GetProvider(),
		),
	}
}

// func (reg *remoteRegistryStatus) ReplicationAPI() globalregistry.ReplicationAPI {
// 	panic("not implemented") // TODO: Implement
// }

func (reg *remoteRegistryStatus) ProjectAPI() globalregistry.RegistryWithProjects {
	panic("not implemented") // TODO: Implement
}

func (reg *remoteRegistryStatus) GetAPIEndpoint() string {
	return reg.Url
}

func (reg *remoteRegistryStatus) GetUsername() string {
	return reg.Credential.AccessKey
}

func (reg *remoteRegistryStatus) GetPassword() string {
	return reg.Credential.AccessSecret
}

func (reg *remoteRegistryStatus) GetName() string {
	return reg.Name
}

func (reg *remoteRegistryStatus) GetProvider() string {
	return reg.Type
}

func (reg *remoteRegistryStatus) GetOptions() globalregistry.RegistryOptions {
	panic("not implemented")
}

func (reg *remoteRegistryStatus) GetAnnotations() map[string]string {
	panic("not implemented")
}

func (r *registry) getRemoteRegistryByNameOrCreate(ctx context.Context, greg globalregistry.Registry) (*remoteRegistryStatus, error) {
	reg, err := r.getRemoteRegistryByName(ctx, greg.GetName())
	if err != nil {
		return nil, err
	}
	if reg == nil {
		reg, err = r.createRemoteRegistry(ctx, greg)
		if err != nil {
			return nil, err
		}
	}
	return reg, nil
}

func (r *registry) getRemoteRegistryByName(ctx context.Context, name string) (*remoteRegistryStatus, error) {
	registries, err := r.listRemoteRegistries(ctx)
	if err != nil {
		return nil, err
	}
	for _, registry := range registries {
		if registry.GetName() == name {
			return registry, nil
		}
	}
	return nil, nil
}

func (r *registry) createRemoteRegistry(ctx context.Context, reg globalregistry.Registry) (*remoteRegistryStatus, error) {
	r.logger.V(1).Info("createRemoteRegistry invoked",
		"reg-name", reg.GetName())
	regStatus := remoteRegistryStatusFromRegistry(reg)
	reqBodyBuf := bytes.NewBuffer(nil)
	err := json.NewEncoder(reqBodyBuf).Encode(regStatus)
	if err != nil {
		return nil, err
	}
	url := *r.parsedUrl
	url.Path = registriesPath
	r.logger.V(1).Info("sending POST request",
		"url", url.String(),
		"body", reqBodyBuf.String(),
	)
	req, err := http.NewRequest(http.MethodPost, url.String(), reqBodyBuf)
	if err != nil {
		return nil, err
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	registryID, err := strconv.Atoi(strings.TrimPrefix(resp.Header.Get("Location"), registriesPath+"/"))
	if err != nil {
		r.logger.Error(err, "cannot parse project ID from response Location header",
			"location-header", resp.Header.Get("Location"))
		return nil, err
	}
	regStatus.Id = registryID
	return regStatus, err
}

func (r *registry) listRemoteRegistries(ctx context.Context) ([]*remoteRegistryStatus, error) {
	url := *r.parsedUrl
	url.Path = registriesPath
	r.logger.V(1).Info("creating new request", "url", url.String())
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	r.logger.V(1).Info("sending HTTP request", "req-uri", req.RequestURI)

	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	registriesResults := []*remoteRegistryStatus{}

	err = json.NewDecoder(resp.Body).Decode(&registriesResults)
	if err != nil {
		r.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		r.logger.Info(b.String())
		fmt.Printf("body: %+v\n", b.String())
	}

	return registriesResults, err
}

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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

const registriesPath = "/api/v2.0/registries"

type remoteRegistries struct {
	reg *registry
}

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

func remoteRegistryStatusFromRegistry(reg globalregistry.RegistryConfig) *remoteRegistryStatus {
	var regType string
	switch reg.GetProvider() {
	case "harbor":
		regType = "harbor"
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
		Insecure: false,
		Type:     regType,
		Description: fmt.Sprintf("%s is a remote %s registry",
			reg.GetName(),
			reg.GetProvider(),
		),
	}
}

func (reg *remoteRegistryStatus) ReplicationAPI() globalregistry.ReplicationAPI {
	panic("not implemented") // TODO: Implement
}

func (reg *remoteRegistryStatus) ProjectAPI() globalregistry.ProjectAPI {
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

func newRemoteRegistries(reg *registry) *remoteRegistries {
	return &remoteRegistries{
		reg: reg,
	}
}

func (r *remoteRegistries) getByNameOrCreate(greg globalregistry.RegistryConfig) (*remoteRegistryStatus, error) {
	reg, err := r.getByName(greg.GetName())
	if err != nil {
		return nil, err
	}
	if reg == nil {
		reg, err = r.create(greg)
		if err != nil {
			return nil, err
		}
	}
	return reg, nil
}

func (r *remoteRegistries) getByName(name string) (*remoteRegistryStatus, error) {
	registries, err := r.list()
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

func (r *remoteRegistries) create(reg globalregistry.RegistryConfig) (*remoteRegistryStatus, error) {
	regStatus := remoteRegistryStatusFromRegistry(reg)
	reqBodyBuf := bytes.NewBuffer(nil)
	err := json.NewEncoder(reqBodyBuf).Encode(regStatus)
	if err != nil {
		return nil, err
	}
	r.reg.logger.V(1).Info(reqBodyBuf.String())
	url := r.reg.parsedUrl
	url.Path = registriesPath
	req, err := http.NewRequest(http.MethodPost, url.String(), reqBodyBuf)
	if err != nil {
		return nil, err
	}

	r.reg.logger.V(1).Info("sending HTTP request", "req-uri", req.RequestURI)

	req.SetBasicAuth(r.reg.GetUsername(), r.reg.GetPassword())

	resp, err := r.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	registryID, err := strconv.Atoi(strings.TrimPrefix(resp.Header.Get("Location"), registriesPath+"/"))
	if err != nil {
		r.reg.logger.Error(err, "cannot parse project ID from response Location header",
			"location-header", resp.Header.Get("Location"))
		return nil, err
	}
	regStatus.Id = registryID
	return regStatus, err
}

func (r *remoteRegistries) list() ([]*remoteRegistryStatus, error) {
	r.reg.parsedUrl.Path = registriesPath
	r.reg.logger.V(1).Info("creating new request", "parsedUrl", r.reg.parsedUrl.String())
	req, err := http.NewRequest(http.MethodGet, r.reg.parsedUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	r.reg.logger.V(1).Info("sending HTTP request", "req-uri", req.RequestURI)

	req.SetBasicAuth(r.reg.GetUsername(), r.reg.GetPassword())

	resp, err := r.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	registriesResults := []*remoteRegistryStatus{}

	err = json.NewDecoder(resp.Body).Decode(&registriesResults)
	if err != nil {
		r.reg.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		r.reg.logger.Info(b.String())
		fmt.Printf("body: %+v\n", b.String())
	}

	return registriesResults, err
}

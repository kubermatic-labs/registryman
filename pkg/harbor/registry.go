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

// harbor package implements the globalregistry.Registry interface for the registry
// provider Harbor.
package harbor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
	logger    logr.Logger
	parsedUrl *url.URL
	globalregistry.Registry
	*http.Client
}

var _ globalregistry.Registry = &registry{}
var _ globalregistry.RegistryWithProjects = &registry{}
var _ globalregistry.ProjectCreator = &registry{}

// newRegistry is the constructor if the registry type. It is a globalregistry RegistryCreator.
func newRegistry(logger logr.Logger, config globalregistry.Registry) (globalregistry.Registry, error) {
	var err error
	c := &registry{
		logger:   logger,
		Registry: config,
		Client:   http.DefaultClient,
	}
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
func (r *registry) do(req *http.Request) (*http.Response, error) {
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

type searchLdapGroupRespBody struct {
	GroupName   string `json:"group_name"`
	LdapGroupDN string `json:"ldap_group_dn"`
}

// searchLdapGroup returns with the distinguished name of the LDAP group. If
// LDAP group is not found, it returns with "", nil.
func (r *registry) searchLdapGroup(ldapGroupName string) (string, error) {
	r.logger.V(1).Info("searching for LDAP group",
		"ldapGroupName", ldapGroupName,
	)
	url := *r.parsedUrl
	url.Path = "/api/v2.0/ldap/groups/search"
	q := url.Query()
	q.Add("groupname", ldapGroupName)
	url.RawQuery = q.Encode()
	r.logger.V(1).Info("sending HTTP request",
		"url", url.String(),
	)
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	parsedResponse := []searchLdapGroupRespBody{}

	err = json.NewDecoder(resp.Body).Decode(&parsedResponse)
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
	switch len(parsedResponse) {
	case 0:
		return "", nil
	case 1:
		return parsedResponse[0].LdapGroupDN, nil
	default:
		return "", fmt.Errorf("multiple LDAP groups found with the name %s", ldapGroupName)
	}
}

func (r *registry) getUserGroups() ([]*userGroup, error) {
	r.logger.V(1).Info("listing usergroups")
	url := *r.parsedUrl
	url.Path = "/api/v2.0/usergroups"
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	parsedResponse := []*userGroup{}

	err = json.NewDecoder(resp.Body).Decode(&parsedResponse)
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
	return parsedResponse, nil
}

func (r *registry) createUserGroup(ug *userGroup) error {
	r.logger.V(1).Info("creating usergroup",
		"GroupName", ug.GroupName,
		"LdapGroupDN", ug.LdapGroupDn,
		"GroupType", ug.GroupType,
	)

	url := *r.parsedUrl
	url.Path = "/api/v2.0/usergroups"
	reqBodyBuf := bytes.NewBuffer(nil)
	err := json.NewEncoder(reqBodyBuf).Encode(ug)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, url.String(), reqBodyBuf)
	if err != nil {
		return err
	}

	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	ug.Id, err = strconv.Atoi(strings.TrimPrefix(resp.Header.Get("Location"), url.Path+"/"))
	if err != nil {
		r.logger.Error(err, "cannot parse project ID from response Location header",
			"location-header", resp.Header.Get("Location"))
		return err
	}

	return nil
}

func (r *registry) deleteUserGroup(ctx context.Context, ug *userGroup) error {
	r.logger.V(1).Info("deleting usergroup",
		"Id", ug.Id,
		"GroupName", ug.GroupName,
		"LdapGroupDN", ug.LdapGroupDn,
		"GroupType", ug.GroupType,
	)

	url := *r.parsedUrl
	url.Path = fmt.Sprintf("/api/v2.0/usergroups/%d", ug.Id)
	req, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	_, err = r.do(req)
	return err
}

func (r *registry) updateIDOfUserGroup(ctx context.Context, ug *userGroup) (bool, error) {
	r.logger.V(1).Info("get id of usergroup",
		"GroupName", ug.GroupName,
		"LdapGroupDN", ug.LdapGroupDn,
		"GroupType", ug.GroupType,
	)
	userGroups, err := r.getUserGroups()
	if err != nil {
		return false, err
	}
	for _, userGroup := range userGroups {
		if ug.GroupName == userGroup.GroupName {
			if ug.GroupType == userGroup.GroupType &&
				ug.LdapGroupDn == userGroup.LdapGroupDn {
				r.logger.V(1).Info("usergroup found",
					"ID", userGroup.Id,
				)
				ug.Id = userGroup.Id
				return true, nil
			}
			r.logger.V(1).Info("usergroup found but with different parameters",
				"LdapGroupDN", userGroup.LdapGroupDn,
				"GroupType", userGroup.GroupType,
			)
			err = r.deleteUserGroup(ctx, userGroup)
			if err != nil {
				return false, err
			}
			return true, r.createUserGroup(ug)
		}
	}
	return false, nil
}

type harborRegistryCapabilities struct{}

func (cap harborRegistryCapabilities) CanPull() bool {
	return true
}

func (cap harborRegistryCapabilities) CanPush() bool {
	return true
}

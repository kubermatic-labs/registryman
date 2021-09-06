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

package jfrog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"strings"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

const projectPath = "/access/api/v1/projects"
const bearerTokenCreationPath = "/artifactory/api/security/token"
const accessTokenCreationPath = "/access/api/v1/tokens"

type metadata struct {
	Severity             string `json:"severity"`
	ReuseSysCVEAllowList string `json:"reuse_sys_cve_allowlist"`
	Public               string `json:"public"`
	PreventVul           string `json:"prevent_vul"`
	EnableContentTrust   string `json:"enable_content_trust"`
	AutoScan             string `json:"auto_scan"`
}

type adminPrivileges struct {
	ManageMembers   bool `json:"manage_members"`
	ManageResources bool `json:"manage_resources"`
}

type projectStatus struct {
	DisplayName                   string          `json:"display_name"`
	Description                   string          `json:"description"`
	AdminPrivileges               adminPrivileges `json:"admin_privileges"`
	StorageQuotaBytes             int             `json:"storage_quota_bytes"`
	SoftLimit                     bool            `json:"soft_limit"`
	StorageQuotaEmailNotification bool            `json:"storage_quota_email_notification"`
	ProjectKey                    string          `json:"project_key"`
}

func (ps *projectStatus) GetName() string {
	return ps.DisplayName
}

type projectCreateReqBody struct {
	Name         string   `json:"project_name"`
	CountLimit   int      `json:"count_limit"`
	RegistryID   int      `json:"registry_id,omitempty"`
	StorageLimit int      `json:"storage_limit"`
	Metadata     metadata `json:"metadata"`
	Public       bool     `json:"public"`
}

type bearerToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type accessToken struct {
	TokenId      string `json:"token_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

func (r *registry) GetProjectByName(name string) (globalregistry.Project, error) {
	if name == "" {
		return &project{
			id:       "",
			registry: r,
			Name:     "",
		}, nil
	}
	projects, err := r.ListProjects()
	if err != nil {
		return nil, err
	}
	for _, project := range projects {
		if project.GetName() == name {
			return project, nil
		}
	}
	return nil, nil
}

func (r *registry) createBearerToken() error {
	r.logger.V(1).Info("creating bearer token",
		"registry", r.GetName(),
	)
	apiUrl := *r.parsedUrl
	apiUrl.Path = bearerTokenCreationPath

	data := url.Values{}
	data.Set("username", r.GetUsername())
	data.Set("scope", "applied-permissions/user")
	data.Set("audience", "jfrt@*")
	data.Set("expire_in", "300")
	data.Set("refreshable", "true")

	req, err := http.NewRequest(http.MethodPost, apiUrl.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	tokenData := &bearerToken{}
	err = json.NewDecoder(resp.Body).Decode(&tokenData)
	if err != nil {
		r.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		r.logger.Info(b.String())
	}

	r.token = tokenData

	return nil
}

func (r *registry) createAccessToken() error {
	r.logger.V(1).Info("creating access token",
		"registry", r.GetName(),
	)
	apiUrl := *r.parsedUrl
	apiUrl.Path = accessTokenCreationPath

	data := url.Values{}
	data.Set("username", r.GetUsername())
	data.Set("scope", "applied-permissions/user")
	data.Set("expire_in", "300")

	req, err := http.NewRequest(http.MethodPost, apiUrl.String(), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header["Content-Type"] = []string{"application/x-www-form-urlencoded"}

	req.Header.Add("Authorization", "Bearer "+r.token.AccessToken)
	// req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	tokenData := &accessToken{}
	err = json.NewDecoder(resp.Body).Decode(&tokenData)
	if err != nil {
		r.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		r.logger.Info(b.String())
	}

	r.aToken = tokenData

	return nil
}

func (r *registry) ListProjects() ([]globalregistry.Project, error) {
	r.logger.V(1).Info("listing projects",
		"registry", r.GetName(),
	)
	apiUrl := *r.parsedUrl
	apiUrl.Path = projectPath
	req, err := http.NewRequest(http.MethodGet, apiUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	r.logger.V(1).Info("token",
		// "token", r.aToken.AccessToken,
		// "scope", r.aToken.Scope,
		"token", r.token.AccessToken,
		"scope", r.token.Scope,
	)
	bearer := "eyJ2ZXIiOiIyIiwidHlwIjoiSldUIiwiYWxnIjoiUlMyNTYiLCJraWQiOiI2bE1hSGxkNXJjM3FCeF9vNG9oLU0xV0FOSXNvQ2xPQV8yRVlpdjBXcnl3In0.eyJleHQiOiJ7XCJyZXZvY2FibGVcIjpcInRydWVcIn0iLCJzdWIiOiJqZmFjQDAxZjVnODJiZ2pqMHh6MG1mcmVhcWYwM2ozXC91c2Vyc1wvdHUtcmVnaXN0cnltYW4iLCJzY3AiOiJhcHBsaWVkLXBlcm1pc3Npb25zXC91c2VyIiwiYXVkIjoiKkAqIiwiaXNzIjoiamZmZUAwMDAiLCJleHAiOjE2NjI0OTM5NDgsImlhdCI6MTYzMDk1Nzk0OCwianRpIjoiMGYyNGRkNGQtZTZkZi00YTdkLWFlNTAtZGRjZWZkZjVjZjU2In0.gRLT8dWKE3y3jXqkZZNgxt4UGgDN49yWO9urHKacAjcOxQWfim9x9qsL9Bb3jVnGLIR4Z9tABNkxmRMP3s2rDduYoKzoq0pYo55haIEBmFGR--eD2dYNmlCngio1Y6F8Eu1pIhoAlk7NpFOJxcAt5UqO_iyndv4-fqpfEDdvoDKC02IZz8dTcv7o28B1Bk4oc26Drk8Pid4v2vFWJuyW9seo1lkzZFk_4zQn79VvxBrY5RLTQGLu16ncyE1yRjaDb0BGbTUPQ7DY3-T2V7_6d6LDOva0p-EspnbN7PBogR46PpqF4KQ9s8kZ5U1C6bX8hK52aC-RtOkm25XSmyqQPQ"
	// req.Header.Add("Authorization", "Bearer "+r.aToken.AccessToken)
	// req.Header.Add("Authorization", "Bearer "+r.token.AccessToken)
	req.Header.Add("Authorization", "Bearer "+bearer)
	req.Header.Add("Accept", "application/json")

	resp, err := r.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	projectData := []*projectStatus{}

	err = json.NewDecoder(resp.Body).Decode(&projectData)
	if err != nil {
		r.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		r.logger.Info(b.String())
	}
	pStatus := make([]globalregistry.Project, len(projectData))
	for i, pData := range projectData {
		r.logger.V(1).Info("listing projects",
			"Name", pData.GetName(),
			"id", pData.ProjectKey,
		)
		pStatus[i] = &project{
			id:       pData.ProjectKey,
			registry: r,
			Name:     pData.GetName(),
		}
	}

	return nil, fmt.Errorf("STOP HERE")
	// return pStatus, err
}

// func (r *registry) CreateProject(name string) (globalregistry.Project, error) {
// 	proj := &project{
// 		registry: r,
// 		Name:     name,
// 	}

// 	apiUrl := *r.parsedUrl
// 	apiUrl.Path = projectPath
// 	reqBodyBuf := bytes.NewBuffer(nil)
// 	err := json.NewEncoder(reqBodyBuf).Encode(&projectCreateReqBody{
// 		Name: proj.Name,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	req, err := http.NewRequest(http.MethodPost, apiUrl.String(), reqBodyBuf)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header["Content-Type"] = []string{"application/json"}
// 	// p.registry.AddBasicAuth(req)
// 	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

// 	resp, err := r.do(req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	defer resp.Body.Close()

// 	projectID, err := strconv.Atoi(strings.TrimPrefix(resp.Header.Get("Location"), projectPath+"/"))
// 	if err != nil {
// 		r.logger.Error(err, "cannot parse project ID from response Location header",
// 			"location-header", resp.Header.Get("Location"))
// 		return nil, err
// 	}
// 	proj.id = projectID

// // Removing default implicit admin user
// members, err := r.getMembers(proj.id)
// if err != nil {
// 	r.logger.V(-1).Info("could not get project members", "error", err)
// 	return proj, nil
// }
// var m *projectMemberEntity
// for _, memb := range members {
// 	if memb.EntityName == r.GetUsername() {
// 		m = memb
// 		break
// 	}
// }
// if m == nil {
// 	r.logger.V(-1).Info("could not find implicit admin member", "username", r.GetUsername())
// 	return proj, nil
// }
// err = r.deleteProjectMember(proj.id, m.Id)
// if err != nil {
// 	r.logger.V(-1).Info("could not delete implicit admin member",
// 		"username", r.GetUsername(),
// 		"error", err,
// 	)
// }
// return proj, nil
// }

func (r *registry) delete(id string) error {
	apiUrl := *r.parsedUrl
	apiUrl.Path = fmt.Sprintf("%s/%d", projectPath, id)
	r.logger.V(1).Info("creating new request", "url", apiUrl.String())
	req, err := http.NewRequest(http.MethodDelete, apiUrl.String(), nil)
	if err != nil {
		return err
	}
	r.logger.V(1).Info("sending HTTP request")

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

type projectRepositoryRespBody struct {
	Name string `json:"name"`
}

func (r *registry) listProjectRepositories(proj *project) ([]string, error) {
	apiUrl := *r.parsedUrl
	apiUrl.Path = fmt.Sprintf("%s/%s/repositories", projectPath, proj.Name)
	req, err := http.NewRequest(http.MethodGet, apiUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	repositories := []*projectRepositoryRespBody{}

	err = json.NewDecoder(resp.Body).Decode(&repositories)
	if err != nil {
		buf := resp.Body.(*bytesBody)
		r.logger.Error(err, "json decoding failed")
		r.logger.Info(buf.String())
	}

	var repositoryNames []string
	for _, rep := range repositories {
		repositoryNames = append(
			repositoryNames,
			strings.TrimPrefix(
				rep.Name,
				proj.Name+"/",
			),
		)
	}
	return repositoryNames, err
}

func (r *registry) deleteProjectRepository(proj *project, repo string) error {
	apiUrl := *r.parsedUrl
	apiUrl.Path = fmt.Sprintf("%s/%s/repositories/%s", projectPath, proj.Name, repo)
	req, err := http.NewRequest(http.MethodDelete, apiUrl.String(), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

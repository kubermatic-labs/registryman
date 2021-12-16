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

package projectbased

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

const projectPath = "/access/api/v1/projects"
const repositoryPath = "/artifactory/api/repositories"

type adminPrivileges struct {
	ManageMembers        bool `json:"manage_members,omitempty"`
	ManageResources      bool `json:"manage_resources,omitempty"`
	ManageSecurityAssets bool `json:"manage_security_assets,omitempty"`
	IndexResources       bool `json:"index_resources,omitempty"`
	AllowIgnoreRules     bool `json:"allow_ignore_rules,omitempty"`
}

type projectStatus struct {
	DisplayName                   string          `json:"display_name"`
	Description                   string          `json:"description,omitempty"`
	AdminPrivileges               adminPrivileges `json:"admin_privileges,omitempty"`
	StorageQuotaBytes             int             `json:"storage_quota_bytes,omitempty"`
	SoftLimit                     bool            `json:"soft_limit,omitempty"`
	StorageQuotaEmailNotification bool            `json:"storage_quota_email_notification,omitempty"`
	ProjectKey                    string          `json:"project_key"`
}

type repositoryConfiguration struct {
	ProjectKey  string `json:"projectKey"`
	Rclass      string `json:"rclass"`
	PackageType string `json:"packageType"`
}

func (ps *projectStatus) GetName() string {
	return ps.DisplayName
}

func (r *projectRegistry) GetProjectByName(ctx context.Context, name string) (globalregistry.Project, error) {
	if name == "" {
		return &project{
			key:      "",
			registry: r,
			Name:     "",
		}, nil
	}
	projects, err := r.ListProjects(ctx)
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

func (r *projectRegistry) ListProjects(ctx context.Context) ([]globalregistry.Project, error) {
	r.logger.V(1).Info("listing projects",
		"registry", r.GetName(),
	)
	apiUrl := *r.parsedUrl
	apiUrl.Path = projectPath
	req, err := http.NewRequest(http.MethodGet, apiUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+r.getAccessToken())
	req.Header.Add("Accept", "application/json")

	resp, err := r.do(ctx, req)
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
			key:      pData.ProjectKey,
			registry: r,
			Name:     pData.GetName(),
		}
	}

	return pStatus, err
}

func (r *projectRegistry) CreateProject(ctx context.Context, name string) (globalregistry.Project, error) {
	r.logger.V(-1).Info("create project",
		"Name", name,
	)
	proj := &project{
		registry: r,
		Name:     name,
	}

	apiUrl := *r.parsedUrl
	apiUrl.Path = projectPath
	reqBodyBuf := bytes.NewBuffer(nil)
	err := json.NewEncoder(reqBodyBuf).Encode(&projectStatus{
		DisplayName: proj.Name,
		ProjectKey:  proj.Name[0:3],
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, apiUrl.String(), reqBodyBuf)
	if err != nil {
		return nil, err
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.Header.Add("Authorization", "Bearer "+r.getAccessToken())
	req.Header.Add("Accept", "application/json")

	resp, err := r.do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	projectData := &projectStatus{}

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
	proj.key = projectData.ProjectKey

	err = r.createRepository(ctx, proj)
	if err != nil {
		return nil, err
	}

	return proj, err
}

func (r *projectRegistry) createRepository(ctx context.Context, proj *project) error {
	r.logger.V(-1).Info("create default docker repository for project",
		"ProjectName", proj.Name,
		"ProjectKey", proj.key,
		"RepoName", proj.GetName()+"-docker",
	)

	apiUrl := *r.parsedUrl
	apiUrl.Path = fmt.Sprintf("%s/%s", repositoryPath, proj.GetName()+"-docker")

	reqBodyBuf := bytes.NewBuffer(nil)
	err := json.NewEncoder(reqBodyBuf).Encode(&repositoryConfiguration{
		ProjectKey:  proj.key,
		Rclass:      "local",
		PackageType: "docker",
	})
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

func (r *projectRegistry) delete(ctx context.Context, key string) error {
	apiUrl := *r.parsedUrl
	apiUrl.Path = fmt.Sprintf("%s/%s", projectPath, key)
	r.logger.V(1).Info("creating new request", "url", apiUrl.String())
	req, err := http.NewRequest(http.MethodDelete, apiUrl.String(), nil)
	if err != nil {
		return err
	}
	r.logger.V(1).Info("sending HTTP request")

	req.Header["Content-Type"] = []string{"application/json"}
	req.Header.Add("Authorization", "Bearer "+r.getAccessToken())
	req.Header.Add("Accept", "application/json")

	resp, err := r.do(ctx, req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

type projectRepositoryRespBody struct {
	Name        string `json:"key"`
	PackageType string `json:"packageType"`
	Type        string `json:"type"`
}

func (r *projectRegistry) listProjectRepositories(ctx context.Context, proj *project) ([]string, error) {
	apiUrl := *r.parsedUrl
	apiUrl.Path = repositoryPath
	req, err := http.NewRequest(http.MethodGet, apiUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(r.GetUsername(), r.GetPassword())
	req.Header.Add("Accept", "application/json")

	resp, err := r.do(ctx, req)
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
		if strings.HasPrefix(rep.Name, fmt.Sprintf("%s-", proj.key)) && rep.PackageType == "Docker" && rep.Type == "LOCAL" {
			repositoryNames = append(
				repositoryNames,
				rep.Name,
			)
		}
	}
	return repositoryNames, err
}

func (r *projectRegistry) deleteProjectRepository(ctx context.Context, proj *project, repo string) error {
	apiUrl := *r.parsedUrl
	apiUrl.Path = fmt.Sprintf("%s/%s", repositoryPath, repo)
	req, err := http.NewRequest(http.MethodDelete, apiUrl.String(), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(ctx, req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

func (r *projectRegistry) getUsedStorage(ctx context.Context, proj *project) (int, error) {
	r.logger.V(1).Info("getting storage usage of a project",
		"projectName", proj.Name,
	)

	apiUrl := *r.parsedUrl
	apiUrl.Path = projectPath
	req, err := http.NewRequest(http.MethodGet, apiUrl.String(), nil)
	if err != nil {
		return -1, err
	}

	req.Header.Add("Authorization", "Bearer "+r.getAccessToken())
	req.Header.Add("Accept", "application/json")

	resp, err := r.do(ctx, req)
	if err != nil {
		return -1, err
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

	for _, pData := range projectData {
		if proj.GetName() == pData.GetName() {
			return pData.StorageQuotaBytes, nil
		}
	}
	return -1, nil
}

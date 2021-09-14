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
	"time"

	"strconv"
	"strings"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

const path = "/api/v2.0/projects"

type metadata struct {
	Severity             string `json:"severity"`
	ReuseSysCVEAllowList string `json:"reuse_sys_cve_allowlist"`
	Public               string `json:"public"`
	PreventVul           string `json:"prevent_vul"`
	EnableContentTrust   string `json:"enable_content_trust"`
	AutoScan             string `json:"auto_scan"`
}

type cveAllowList struct {
	CreationTime time.Time     `json:"creation_time"`
	ExpiresAt    int64         `json:"expires_at"`
	UpdateTime   time.Time     `json:"update_time"`
	ID           int           `json:"id"`
	ProjectID    int           `json:"project_id"`
	Items        []interface{} `json:"items"`
}

type projectStatus struct {
	CreationTime       time.Time    `json:"creation_time" yaml:"creation_time"`
	UpdateTime         time.Time    `json:"update_time"`
	ProjectID          int          `json:"project_id"`
	OwnerID            int          `json:"owner_id"`
	Name               string       `json:"name"`
	Metadata           metadata     `json:"metadata"`
	CVEAllowList       cveAllowList `json:"cve_allowlist"`
	CurrentUserRoleIDs []int        `json:"current_user_role_ids"`
	CurrentUserRoleID  int          `json:"current_user_role_id"`
}

func (ps *projectStatus) GetName() string {
	return ps.Name
}

type projectCreateReqBody struct {
	Name         string       `json:"project_name"`
	CVEAllowList cveAllowList `json:"cve_allowlist"`
	CountLimit   int          `json:"count_limit"`
	RegistryID   int          `json:"registry_id,omitempty"`
	StorageLimit int          `json:"storage_limit"`
	Metadata     metadata     `json:"metadata"`
	Public       bool         `json:"public"`
}

func (r *registry) GetProjectByName(ctx context.Context, name string) (globalregistry.Project, error) {
	if name == "" {
		return &project{
			id:       -1,
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

func (r *registry) ListProjects(ctx context.Context) ([]globalregistry.Project, error) {
	r.logger.V(1).Info("listing projects",
		"registry", r.GetName(),
	)
	url := *r.parsedUrl
	url.Path = path
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

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
		pStatus[i] = &project{
			id:       pData.ProjectID,
			registry: r,
			Name:     pData.Name,
		}
	}
	return pStatus, err
}

func (r *registry) CreateProject(ctx context.Context, name string) (globalregistry.Project, error) {
	proj := &project{
		registry: r,
		Name:     name,
	}

	url := *r.parsedUrl
	url.Path = path
	reqBodyBuf := bytes.NewBuffer(nil)
	err := json.NewEncoder(reqBodyBuf).Encode(&projectCreateReqBody{
		Name: proj.Name,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url.String(), reqBodyBuf)
	if err != nil {
		return nil, err
	}
	req.Header["Content-Type"] = []string{"application/json"}
	// p.registry.AddBasicAuth(req)
	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	projectID, err := strconv.Atoi(strings.TrimPrefix(resp.Header.Get("Location"), path+"/"))
	if err != nil {
		r.logger.Error(err, "cannot parse project ID from response Location header",
			"location-header", resp.Header.Get("Location"))
		return nil, err
	}
	proj.id = projectID

	// Removing default implicit admin user
	members, err := r.getMembers(ctx, proj.id)
	if err != nil {
		r.logger.V(-1).Info("could not get project members", "error", err)
		return proj, nil
	}
	var m *projectMemberEntity
	for _, memb := range members {
		if memb.EntityName == r.GetUsername() {
			m = memb
			break
		}
	}
	if m == nil {
		r.logger.V(-1).Info("could not find implicit admin member", "username", r.GetUsername())
		return proj, nil
	}
	err = r.deleteProjectMember(ctx, proj.id, m.Id)
	if err != nil {
		r.logger.V(-1).Info("could not delete implicit admin member",
			"username", r.GetUsername(),
			"error", err,
		)
	}
	return proj, nil
}

func (r *registry) delete(ctx context.Context, id int) error {
	url := *r.parsedUrl
	url.Path = fmt.Sprintf("%s/%d", path, id)
	r.logger.V(1).Info("creating new request", "url", url.String())
	req, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	if err != nil {
		return err
	}
	r.logger.V(1).Info("sending HTTP request")

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(ctx, req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

type projectRepositoryRespBody struct {
	Name string `json:"name"`
}

func (r *registry) listProjectRepositories(ctx context.Context, proj *project) ([]string, error) {
	url := *r.parsedUrl
	url.Path = fmt.Sprintf("%s/%s/repositories", path, proj.Name)
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

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

func (r *registry) deleteProjectRepository(ctx context.Context, proj *project, repo string) error {
	url := *r.parsedUrl
	url.Path = fmt.Sprintf("%s/%s/repositories/%s", path, proj.Name, repo)
	req, err := http.NewRequest(http.MethodDelete, url.String(), nil)
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

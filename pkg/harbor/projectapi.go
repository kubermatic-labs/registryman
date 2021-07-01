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

type projectAPI struct {
	reg *registry
}

func newProjectAPI(reg *registry) (*projectAPI, error) {
	return &projectAPI{
		reg: reg,
	}, nil
}

func (p *projectAPI) GetByName(name string) (globalregistry.Project, error) {
	projects, err := p.List()
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

func (p *projectAPI) List() ([]globalregistry.Project, error) {
	p.reg.logger.V(1).Info("listing projects",
		"registry", p.reg.GetName(),
	)
	url := *p.reg.parsedUrl
	url.Path = path
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(p.reg.GetUsername(), p.reg.GetPassword())

	resp, err := p.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	projectData := []*projectStatus{}

	err = json.NewDecoder(resp.Body).Decode(&projectData)
	if err != nil {
		p.reg.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		p.reg.logger.Info(b.String())
	}
	pStatus := make([]globalregistry.Project, len(projectData))
	for i, pData := range projectData {
		pStatus[i] = &project{
			id:   pData.ProjectID,
			api:  p,
			Name: pData.Name,
			sApi: newScannerAPI(p.reg),
		}
	}
	return pStatus, err
}

func (p *projectAPI) Create(name string) (globalregistry.Project, error) {
	proj := &project{
		api:  p,
		Name: name,
		sApi: newScannerAPI(p.reg),
	}

	url := *p.reg.parsedUrl
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
	req.SetBasicAuth(p.reg.GetUsername(), p.reg.GetPassword())

	resp, err := p.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	projectID, err := strconv.Atoi(strings.TrimPrefix(resp.Header.Get("Location"), path+"/"))
	if err != nil {
		p.reg.logger.Error(err, "cannot parse project ID from response Location header",
			"location-header", resp.Header.Get("Location"))
		return nil, err
	}
	proj.id = projectID

	// Removing default implicit admin user
	members, err := p.getMembers(proj.id)
	if err != nil {
		p.reg.logger.V(-1).Info("could not get project members", "error", err)
		return proj, nil
	}
	var m *projectMemberEntity
	for _, memb := range members {
		if memb.EntityName == p.reg.GetUsername() {
			m = memb
			break
		}
	}
	if m == nil {
		p.reg.logger.V(-1).Info("could not find implicit admin member", "username", p.reg.GetUsername())
		return proj, nil
	}
	err = p.deleteProjectMember(proj.id, m.Id)
	if err != nil {
		p.reg.logger.V(-1).Info("could not delete implicit admin member",
			"username", p.reg.GetUsername(),
			"error", err,
		)
	}
	return proj, nil
}

func (p *projectAPI) delete(id int) error {
	url := *p.reg.parsedUrl
	url.Path = fmt.Sprintf("%s/%d", path, id)
	p.reg.logger.V(1).Info("creating new request", "url", url.String())
	req, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	if err != nil {
		return err
	}
	p.reg.logger.V(1).Info("sending HTTP request")

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(p.reg.GetUsername(), p.reg.GetPassword())

	resp, err := p.reg.do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

type projectRepositoryRespBody struct {
	Name string `json:"name"`
	proj *project
}

func (p *projectAPI) listProjectRepositories(proj *project) ([]*projectRepositoryRespBody, error) {
	url := *p.reg.parsedUrl
	url.Path = fmt.Sprintf("%s/%s/repositories", path, proj.Name)
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(p.reg.GetUsername(), p.reg.GetPassword())

	resp, err := p.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	repositories := []*projectRepositoryRespBody{}

	err = json.NewDecoder(resp.Body).Decode(&repositories)
	if err != nil {
		buf := resp.Body.(*bytesBody)
		p.reg.logger.Error(err, "json decoding failed")
		p.reg.logger.Info(buf.String())
	}
	for _, rep := range repositories {
		rep.proj = proj
		rep.Name = strings.TrimPrefix(
			rep.Name,
			proj.Name+"/")
	}
	return repositories, err
}

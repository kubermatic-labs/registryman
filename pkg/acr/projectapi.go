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
package acr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type projectAPI struct {
	reg *registry
}

func newProjectAPI(reg *registry) (*projectAPI, error) {
	return &projectAPI{
		reg: reg,
	}, nil
}

func (p *projectAPI) Create(name string) (globalregistry.Project, error) {
	return nil, fmt.Errorf("not implemented")
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

type bytesBody struct {
	*bytes.Buffer
}

func (bb bytesBody) Close() error { return nil }

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

func (p *projectAPI) List() ([]globalregistry.Project, error) {
	p.reg.parsedUrl.Path = path
	req, err := http.NewRequest(http.MethodGet, p.reg.parsedUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(p.reg.GetUsername(), p.reg.GetPassword())

	resp, err := p.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	projectData := &repositories{}

	fmt.Println(resp.Body)
	err = json.NewDecoder(resp.Body).Decode(projectData)
	if err != nil {
		p.reg.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		p.reg.logger.Info(b.String())
	}

	fmt.Println(*projectData)
	pStatus := p.collectProjectNamesFromRepos(projectData.Repositories)

	return pStatus, err
}

func (p *projectAPI) collectProjectNamesFromRepos(repoNames []string) []globalregistry.Project {
	projectNames := make(map[string]struct{})

	for _, pData := range repoNames {
		pName := strings.Split(pData, "/")[0]
		projectNames[pName] = struct{}{}
	}
	pStatus := make([]globalregistry.Project, len(projectNames))

	i := 0
	for projectName := range projectNames {
		pStatus[i] = &project{
			api:  p,
			name: projectName,
		}
		i++
	}
	return pStatus
}

//func (p *projectAPI) delete(name string) error {
//	return fmt.Errorf("not implemented")
//}

//func (p *projectAPI) listProjectRepositories(proj *project) ([]globalregistry.Repository, error) {
//	return nil, fmt.Errorf("not implemented")
//}

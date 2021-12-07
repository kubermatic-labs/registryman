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
	"strings"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

const artifactoryPath = "/artifactory"
const permissionPath = artifactoryPath + "/api/security/permissions"
const storagePath = artifactoryPath + "/api/storage"

type project struct {
	name     string
	registry *pathRegistry
}

type folderInfoRespBody struct {
	Uri      string      `json:"uri"`
	Repo     string      `json:"repo"`
	Path     string      `json:"path"`
	Children []childItem `json:"children"`
}

type childItem struct {
	Uri      string `json:"uri"`
	IsFolder bool   `json:"folder"`
}

func (r *pathRegistry) GetProjectByName(ctx context.Context, name string) (globalregistry.Project, error) {
	if name == "" {
		return &project{
			name:     "",
			registry: r,
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
	return nil, fmt.Errorf("no project found: %w", globalregistry.ErrRecoverableError)
}

func (r *pathRegistry) listFolders(ctx context.Context, optionalProjectName string) ([]string, error) {
	apiUrl := *r.parsedUrl
	path := storagePath + "/" + r.GetDockerRegistryName()
	if optionalProjectName != "" {
		path += "/" + optionalProjectName
	}
	apiUrl.Path = path
	req, err := http.NewRequest(http.MethodGet, apiUrl.String(), nil)
	r.logger.V(1).Info("creating new request", "url", apiUrl.String())
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	repos := &folderInfoRespBody{}

	err = json.NewDecoder(resp.Body).Decode(repos)
	if err != nil {
		r.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		r.logger.Info(b.String())
	}
	folderNames := []string{}

	for _, item := range repos.Children {
		if item.IsFolder {
			folderNames = append(folderNames, strings.TrimPrefix(item.Uri, "/"))
		}

	}

	return folderNames, nil
}

func (r *pathRegistry) ListProjects(ctx context.Context) ([]globalregistry.Project, error) {
	repositories, err := r.listFolders(ctx, "")
	if err != nil {
		return nil, err
	}
	pStatus := r.collectProjectNamesFromRepos(repositories)

	return pStatus, err
}

func projectNameFromRepoName(repoName string) string {
	return strings.Split(repoName, "/")[0]
}

func (r *pathRegistry) collectProjectNamesFromRepos(repoNames []string) []globalregistry.Project {
	projectNames := make(map[string]struct{})

	for _, repoName := range repoNames {
		projectName := projectNameFromRepoName(repoName)
		projectNames[projectName] = struct{}{}
	}
	pStatus := make([]globalregistry.Project, len(projectNames))

	i := 0
	for projectName := range projectNames {
		pStatus[i] = &project{
			name:     projectName,
			registry: r,
		}
		i++
	}
	return pStatus
}

func (r *pathRegistry) CreateProject(ctx context.Context, name string) (globalregistry.Project, error) {
	proj := &project{
		registry: r,
		name:     name,
	}

	apiUrl := *r.parsedUrl
	apiUrl.Path = fmt.Sprintf("%s/%s/%s/", artifactoryPath, r.GetDockerRegistryName(), proj.GetName())

	req, err := http.NewRequest(http.MethodPut, apiUrl.String(), nil)
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

	err = r.createPermission(ctx, name, &permissionConfiguration{
		Name:            r.GetDockerRegistryName() + "_" + name,
		IncludesPattern: name + "/**",
		Repositories:    []string{r.GetDockerRegistryName()},
	})

	if err != nil {
		return nil, err
	}

	return proj, nil
}

func (r *pathRegistry) delete(ctx context.Context, project string) error {
	apiUrl := *r.parsedUrl
	apiUrl.Path = fmt.Sprintf("%s/%s/%s", artifactoryPath, r.GetDockerRegistryName(), project)
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

	err = r.deletePermission(ctx, project)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("principal target %s not exists, %w", r.GetDockerRegistryName()+"_"+project, globalregistry.ErrRecoverableError)
		}
	}

	return err
}

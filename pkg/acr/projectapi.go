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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

func (r *registry) GetProjectByName(ctx context.Context, name string) (globalregistry.Project, error) {
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

	switch {
	case resp.StatusCode == 401:
		// Unauthorized
		return nil, globalregistry.ErrUnauthorized
	case resp.StatusCode < 200 || resp.StatusCode >= 300:
		// Any other error code
		s.logger.V(-1).Info("HTTP response status code is not OK",
			"status-code", resp.StatusCode,
			"resp-body-size", n,
			"req-url", req.URL,
		)
		s.logger.V(1).Info(buf.String())
	}
	return resp, nil
}

func (r *registry) getRepositories(ctx context.Context) ([]string, error) {
	r.parsedUrl.Path = path
	req, err := http.NewRequest(http.MethodGet, r.parsedUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	resp, err := r.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	repos := &repositories{}

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
	return repos.Repositories, nil
}

func (r *registry) ListProjects(ctx context.Context) ([]globalregistry.Project, error) {
	repositories, err := r.getRepositories(ctx)
	if err != nil {
		return nil, err
	}
	pStatus := r.collectProjectNamesFromRepos(repositories)

	return pStatus, err
}

func projectNameFromRepoName(repoName string) string {
	return strings.Split(repoName, "/")[0]
}

func (r *registry) collectProjectNamesFromRepos(repoNames []string) []globalregistry.Project {
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

func collectReposOfProject(projectName string, repoNames []string) []string {
	reposOfProject := []string{}
	for _, repoName := range repoNames {
		if projectNameFromRepoName(repoName) == projectName {
			reposOfProject = append(reposOfProject, repoName)
		}
	}
	return reposOfProject
}

func (r *registry) deleteRepoOfProject(repoName string) error {
	r.logger.V(1).Info("deleting ACR repository",
		"repositoryName", repoName,
	)
	url := *r.parsedUrl
	url.Path = fmt.Sprintf("/acr/v1/%s", repoName)
	req, err := http.NewRequest(http.MethodDelete, url.String(), nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(r.GetUsername(), r.GetPassword())

	_, err = r.do(req)
	return err
}

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
	"context"
	"fmt"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type project struct {
	name     string
	registry *registry
}

var _ globalregistry.Project = &project{}
var _ globalregistry.DestructibleProject = &project{}
var _ globalregistry.ProjectWithRepositories = &project{}

func (p *project) GetName() string {
	return p.name
}

// Delete implements the globalregistry.Project interface. It succeeds of there
// are no repos of the projects. Otherwise, it returns an error.
//
func (p *project) Delete(ctx context.Context) error {
	repoNames, err := p.registry.getRepositories(ctx)
	if err != nil {
		return err
	}
	reposOfProject := collectReposOfProject(p.name, repoNames)
	if len(reposOfProject) != 0 {
		switch opt := p.registry.GetOptions().(type) {
		case globalregistry.CanForceDelete:
			if !opt.ForceDeleteProjects() {
				return fmt.Errorf("%s: repositories are present, please delete them before deleting the project, %w", p.GetName(), globalregistry.ErrRecoverableError)
			}
			for _, repo := range repoNames {
				p.registry.logger.V(1).Info("deleting repository",
					"repositoryName", repoNames,
				)
				err = p.deleteRepository(repo)
				if err != nil {
					return err
				}
			}
		default:
			return globalregistry.ErrNotImplemented
		}
	}
	return nil
}

func (p *project) GetRepositories(ctx context.Context) ([]string, error) {
	repos, err := p.registry.getRepositories(ctx)
	if err != nil {
		return nil, err
	}
	return collectReposOfProject(p.name, repos), nil
}

func (p *project) deleteRepository(repoName string) error {
	return p.registry.deleteRepoOfProject(p, repoName)
}

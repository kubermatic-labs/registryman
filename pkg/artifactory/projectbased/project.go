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
	"context"
	"fmt"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type project struct {
	key      string
	registry *projectRegistry
	Name     string
}

// interface guard
var _ globalregistry.Project = &project{}
var _ globalregistry.ProjectWithRepositories = &project{}
var _ globalregistry.ProjectWithMembers = &project{}

// var _ globalregistry.MemberManipulatorProject = &project{}
var _ globalregistry.DestructibleProject = &project{}

func (p *project) GetName() string {
	return p.Name
}

// Delete removes the project from registry
func (p *project) Delete(ctx context.Context) error {
	repos, err := p.GetRepositories(ctx)
	if err != nil {
		return err
	}

	if len(repos) > 0 {
		switch opt := p.registry.GetOptions().(type) {
		case globalregistry.CanForceDelete:
			if f := opt.ForceDeleteProjects(); !f {
				return fmt.Errorf("%s: repositories are present, please delete them before deleting the project, %w", p.Name, globalregistry.ErrRecoverableError)
			}
			for _, repo := range repos {
				p.registry.logger.V(1).Info("deleting repository",
					"repositoryName", repos,
				)
				err = p.deleteRepository(ctx, repo)
				if err != nil {
					return err
				}
			}
		}

	}
	return p.registry.delete(ctx, p.key)
}

func (p *project) GetMembers(ctx context.Context) ([]globalregistry.ProjectMember, error) {
	members, err := p.registry.getMembers(ctx, p.key)
	if err != nil {
		return nil, err
	}
	projectMembers := make([]globalregistry.ProjectMember, len(members))

	c := 0
	for _, m := range members {
		projectMembers[c] = m.toProjectMember()
		c++
	}

	return projectMembers, nil
}

func (p *project) GetRepositories(ctx context.Context) ([]string, error) {
	return p.registry.listProjectRepositories(ctx, p)
}

func (p *project) deleteRepository(ctx context.Context, r string) error {
	return p.registry.deleteProjectRepository(ctx, p, r)
}

// GetUsedStorage implements the globalregistry.Project interface.
func (p *project) GetUsedStorage(ctx context.Context) (int, error) {
	return p.registry.getUsedStorage(ctx, p)
}

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
	"fmt"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type project struct {
	name string
	api  *projectAPI
}

var _ globalregistry.Project = &project{}

func (p *project) GetName() string {
	return p.name
}

// Delete implements the globalregistry.Project interface. It succeeds of there
// are no repos of the projects. Otherwise, it returns an error.
//
func (p *project) Delete() error {
	repoNames, err := p.api.getRepositories()
	if err != nil {
		return err
	}
	reposOfProject := p.api.collectReposOfProject(p.name, repoNames)
	if len(reposOfProject) != 0 {
		switch opt := p.api.reg.GetOptions().(type) {
		case globalregistry.CanForceDelete:
			if !opt.ForceDeleteProjects() {
				return fmt.Errorf("%s: repositories are present, please delete them before deleting the project, %w", p.GetName(), globalregistry.ErrRecoverableError)
			}
			for _, repo := range repoNames {
				p.api.reg.logger.V(1).Info("deleting repository",
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

// AssignMember implements the globalregistry.Project interface. Currently, it
// is not implemented.
func (p *project) AssignMember(member globalregistry.ProjectMember) (*globalregistry.ProjectMemberCredentials, error) {
	return nil, fmt.Errorf("cannot assign member to a project in ACR: %w",
		globalregistry.ErrNotImplemented)
}

// UnassignMember implements the globalregistry.Project interface. Currently, it
// is not implemented.
func (p *project) UnassignMember(member globalregistry.ProjectMember) error {
	return fmt.Errorf("cannot assign member to a project in ACR: %w",
		globalregistry.ErrNotImplemented)
}

// AssignReplicationRule implements the globalregistry.Project interface.
// Currently, it is not implemented.
func (p *project) AssignReplicationRule(remoteReg globalregistry.RegistryConfig, trigger, direction string) (globalregistry.ReplicationRule, error) {
	return nil, fmt.Errorf("cannot assign the replication rule to a project in ACR: %w", globalregistry.ErrNotImplemented)
}

func (p *project) GetMembers() ([]globalregistry.ProjectMember, error) {
	p.api.reg.logger.V(-1).Info("ACR.GetMembers not implemented")
	return []globalregistry.ProjectMember{}, nil
}

func (p *project) GetReplicationRules(trigger, direction string) ([]globalregistry.ReplicationRule, error) {

	return nil, nil
}

func (p *project) AssignScanner(s globalregistry.Scanner) error {
	return fmt.Errorf("method ACR.AssignScanner not implemented: %w", globalregistry.ErrRecoverableError)
}

func (p *project) GetScanner() (globalregistry.Scanner, error) {
	return nil, nil
}

func (p *project) UnassignScanner(s globalregistry.Scanner) error {
	return fmt.Errorf("method ACR.UnassignScanner not implemented: %w", globalregistry.ErrRecoverableError)
}

func (p *project) GetRepositories() ([]string, error) {
	return p.api.getRepositories()
}

// GetUsedStorage implements the globalregistry.Project interface. Currently, it
// is not implemented.
func (p *project) GetUsedStorage() (int, error) {
	return -1, fmt.Errorf("cannot get used storage of a project in ACR: %w",
		globalregistry.ErrNotImplemented)
}

func (p *project) deleteRepository(repoName string) error {
	return p.api.deleteRepoOfProject(p, repoName)
}

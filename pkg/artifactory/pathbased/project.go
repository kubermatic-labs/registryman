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
	"context"
	"fmt"
	"strings"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

// interface guard
var _ globalregistry.Project = &project{}
var _ globalregistry.ProjectWithRepositories = &project{}
var _ globalregistry.ProjectWithMembers = &project{}

var _ globalregistry.MemberManipulatorProject = &project{}
var _ globalregistry.DestructibleProject = &project{}

func (p *project) GetName() string {
	return p.name
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
				return fmt.Errorf("%s: repositories are present, please delete them before deleting the project, %w", p.GetName(), globalregistry.ErrRecoverableError)
			}
		}

	}
	return p.registry.delete(ctx, p.GetName())
}

func (p *project) AssignMember(ctx context.Context, member globalregistry.ProjectMember) (*globalregistry.ProjectMemberCredentials, error) {
	role, err := roleFromString(member.GetRole())
	if err != nil {
		return nil, err
	}
	permissionReqBody, err := p.registry.getPermission(ctx, p.registry.GetDockerRegistryName()+"_"+p.GetName())
	if err != nil {
		return nil, err
	}

	switch member.GetType() {
	default:
		panic(fmt.Sprintf("unhandled ProjectMemberType: %s", member.GetType()))
	case userType:
		if permissionReqBody.Principals.Users == nil {
			permissionReqBody.Principals.Users = make(map[string][]string)
		}

		permissionReqBody.Principals.Users[member.GetName()] = strings.Split(role.String(), ",")

	case groupType:
		if permissionReqBody.Principals.Groups == nil {
			permissionReqBody.Principals.Groups = make(map[string][]string)
		}

		permissionReqBody.Principals.Groups[member.GetName()] = strings.Split(role.String(), ",")
	}

	err = p.registry.createPermission(ctx, p.GetName(), permissionReqBody)
	return nil, err

}

func (p *project) GetMembers(ctx context.Context) ([]globalregistry.ProjectMember, error) {
	members, err := p.registry.getMembers(ctx, p)
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (p *project) UnassignMember(ctx context.Context, member globalregistry.ProjectMember) error {

	var m globalregistry.ProjectMember
	members, err := p.registry.getMembers(ctx, p)
	if err != nil {
		return err
	}
	permissionReqBody, err := p.registry.getPermission(ctx, p.registry.GetDockerRegistryName()+"_"+p.GetName())
	if err != nil {
		return err
	}
	switch member.GetType() {
	case userType:
		for _, memb := range members {
			switch memb.(type) {
			case projectMember:
				if memb.GetName() == member.GetName() {
					m = memb
					break
				}
			}
		}
		if m == nil {
			return fmt.Errorf("user member not found")
		}

		delete(permissionReqBody.Principals.Users, m.GetName())
	case groupType:
		for _, memb := range members {
			switch memb.(type) {
			case groupMember:
				if memb.GetName() == member.GetName() {
					m = memb
					break
				}
			}
		}
		if m == nil {
			return fmt.Errorf("group member not found")
		}

		delete(permissionReqBody.Principals.Groups, m.GetName())
	}

	err = p.registry.createPermission(ctx, p.GetName(), permissionReqBody)
	return err
}

func (p *project) GetRepositories(ctx context.Context) ([]string, error) {
	repos, err := p.registry.listFolders(ctx, p.GetName())
	if err != nil {
		return nil, err
	}
	return repos, nil
}

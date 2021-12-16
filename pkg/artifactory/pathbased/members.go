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

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

const (
	userType  = "User"
	groupType = "Group"
)

type projectMemberEntity struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

type projectMember projectMemberEntity

var _ globalregistry.ProjectMember = &projectMember{}

func (m projectMember) GetName() string {
	return m.Name
}

func (m projectMember) GetType() string {
	return userType
}

func (m projectMember) GetRole() string {
	return roleFromList(m.Roles)
}

type groupMember projectMemberEntity

var _ globalregistry.ProjectMember = &groupMember{}

func (m groupMember) GetName() string {
	return m.Name
}

func (m groupMember) GetType() string {
	return groupType
}

func (m groupMember) GetRole() string {
	return roleFromList(m.Roles)
}

func (r *pathRegistry) getMembers(ctx context.Context, p *project) ([]globalregistry.ProjectMember, error) {
	projectMembers, err := p.registry.getPermission(ctx, r.GetDockerRegistryName()+"_"+p.GetName())
	if err != nil {
		return nil, err
	}

	projectMembersResult := make([]globalregistry.ProjectMember, len(projectMembers.Principals.Users)+len(projectMembers.Principals.Groups))

	c := 0
	for user, roles := range projectMembers.Principals.Users {
		projectMembersResult[c] = projectMember{
			Name:  user,
			Roles: roles,
		}
		c++
	}
	for group, roles := range projectMembers.Principals.Groups {
		projectMembersResult[c] = groupMember{
			Name:  group,
			Roles: roles,
		}
		c++
	}

	return projectMembersResult, err
}

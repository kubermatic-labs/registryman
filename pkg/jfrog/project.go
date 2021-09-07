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

package jfrog

import (
	"fmt"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type project struct {
	key      string
	registry *registry
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
func (p *project) Delete() error {
	repos, err := p.GetRepositories()
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
				err = p.deleteRepository(repo)
				if err != nil {
					return err
				}
			}
		}

	}
	return p.registry.delete(p.key)
}

// func robotRoleToAccess(role string) []access {
// 	switch role {
// 	case "PushOnly":
// 		return []access{
// 			{
// 				Action:   "push",
// 				Resource: "repository",
// 				// Effect:   "",
// 			},
// 		}
// 	case "PullOnly":
// 		return []access{
// 			{
// 				Action:   "pull",
// 				Resource: "repository",
// 				// Effect:   "",
// 			},
// 		}
// 	case "PullAndPush":
// 		return []access{
// 			{
// 				Action:   "pull",
// 				Resource: "repository",
// 				// Effect:   "",
// 			},
// 			{
// 				Action:   "push",
// 				Resource: "repository",
// 				// Effect:   "",
// 			},
// 		}
// 	default:
// 		panic(fmt.Sprintf("%s robot role is not supported", role))
// 	}
// }

// func (p *project) AssignMember(member globalregistry.ProjectMember) (*globalregistry.ProjectMemberCredentials, error) {
// 	role, err := roleFromString(member.GetRole())
// 	if err != nil {
// 		return nil, err
// 	}
// 	pum := &projectMemberRequestBody{
// 		RoleId: role,
// 		MemberUser: &userEntity{
// 			Username: member.GetName(),
// 		},
// 	}
// 	_, err = p.registry.createProjectMember(p.id, pum)
// 	return nil, err

// }

func (p *project) GetMembers() ([]globalregistry.ProjectMember, error) {
	members, err := p.registry.getMembers(p.key)
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

// func (p *project) UnassignMember(member globalregistry.ProjectMember) error {
// 	memberType := member.GetType()
// 	var err error
// 	switch memberType {
// 	default:
// 		return fmt.Errorf("unhandled member type: %s", memberType)
// 	case userType, groupType:
// 		var m *projectMemberEntity
// 		var members []*projectMemberEntity
// 		members, err = p.registry.getMembers(p.id)
// 		if err != nil {
// 			return err
// 		}
// 		for _, memb := range members {
// 			if memb.EntityName == member.GetName() {
// 				m = memb
// 				break
// 			}
// 		}
// 		if m == nil {
// 			return fmt.Errorf("user member not found")
// 		}
// 		err = p.registry.deleteProjectMember(p.id, m.Id)
// 	case robotType:
// 		var m *robot
// 		var members []*robot
// 		members, err = p.registry.getRobotMembers(p.id)
// 		if err != nil {
// 			return err
// 		}
// 		expectedName := fmt.Sprintf("robot$%s+%s", p.GetName(), member.GetName())
// 		for _, memb := range members {
// 			if memb.GetName() == expectedName {
// 				m = memb
// 				break
// 			}
// 		}
// 		if m == nil {
// 			return fmt.Errorf("robot member not found")
// 		}
// 		err = p.registry.deleteProjectRobotMember(p.id, m.Id)
// 	}
// 	return err
// }

// func (p *project) AssignReplicationRule(remoteReg globalregistry.Registry, trigger, direction string) (globalregistry.ReplicationRule, error) {
// 	return p.registry.createReplicationRule(p, remoteReg, trigger, direction)
// }

func (p *project) GetRepositories() ([]string, error) {
	return p.registry.listProjectRepositories(p)
}

func (p *project) deleteRepository(r string) error {
	return p.registry.deleteProjectRepository(p, r)
}

// GetUsedStorage implements the globalregistry.Project interface.
func (p *project) GetUsedStorage() (int, error) {
	return p.registry.getUsedStorage(p)
}

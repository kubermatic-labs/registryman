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
	"fmt"
	"strings"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type project struct {
	id   int
	api  *projectAPI
	Name string
}

// interface guard
var _ globalregistry.Project = &project{}

func (p *project) GetName() string {
	return p.Name
}

// Delete removes the project from registry
func (p *project) Delete() error {
	repos, err := p.getRepositories()
	if err != nil {
		return err
	}
	if len(repos) > 0 {
		return fmt.Errorf("%s: repositories are present, please delete them before deleting the project, %w", p.Name, globalregistry.RecoverableError)
	}
	return p.api.delete(p.id)
}

func robotRoleToAccess(role string) []access {
	switch role {
	case "PushOnly":
		return []access{
			{
				Action:   "push",
				Resource: "repository",
				// Effect:   "",
			},
		}
	case "PullOnly":
		return []access{
			{
				Action:   "pull",
				Resource: "repository",
				// Effect:   "",
			},
		}
	case "PullAndPush":
		return []access{
			{
				Action:   "pull",
				Resource: "repository",
				// Effect:   "",
			},
			{
				Action:   "push",
				Resource: "repository",
				// Effect:   "",
			},
		}
	default:
		panic(fmt.Sprintf("%s robot role is not supported", role))
	}
}

func (p *project) AssignMember(member globalregistry.ProjectMember) (*globalregistry.ProjectMemberCredentials, error) {
	memberType := member.GetType()
	switch memberType {
	default:
		return nil, fmt.Errorf("unhandled member type: %s", memberType)
	case "User":
		role, err := roleFromString(member.GetRole())
		if err != nil {
			return nil, err
		}
		pum := &projectMemberRequestBody{
			RoleId: role,
			MemberUser: &userEntity{
				Username: member.GetName(),
			},
		}
		_, err = p.api.createProjectMember(p.id, pum)
		if err != nil {
			return nil, err
		}
		return nil, nil
	case "Robot":
		prm := &robot{
			Description: "generated robot member",
			Level:       "project",
			// Editable:     false,
			// ExpiresAt:    0,
			Name: member.GetName(),
			// Disable:      false,
			// Duration:     0,
			// Id:           0,
			Permissions: []robotPermission{
				{
					Access:    robotRoleToAccess(member.GetRole()),
					Kind:      "project",
					Namespace: p.GetName(),
				},
			},
		}
		// Name:        member.GetName(),
		// ExpiresAt:   1024,
		// Description: "generated robot member",
		// Access:      robotRoleToAccess(member.GetRole()),
		r, err := p.api.createProjectRobotMember(prm)
		if err != nil {
			return nil, err
		}
		return &globalregistry.ProjectMemberCredentials{
			Username: r.Name,
			Password: r.Secret,
		}, nil
	}

}

func (p *project) GetMembers() ([]globalregistry.ProjectMember, error) {
	userMembers, err := p.api.getUserMembers(p.id)
	if err != nil {
		return nil, err
	}
	robotMembers, err := p.api.getRobotMembers(p.id)
	if err != nil {
		return nil, err
	}
	members := make([]globalregistry.ProjectMember, len(userMembers)+len(robotMembers))

	c := 0
	// collecting the members of type User
	for _, user := range userMembers {
		members[c] = user
		c++
	}

	// collecting the members of type Robot
	for _, robot := range robotMembers {
		robot.Name = strings.TrimPrefix(robot.Name, fmt.Sprintf("robot$%s+", p.GetName()))
		members[c] = robot
		c++
	}
	return members, nil
}

func (p *project) UnassignMember(member globalregistry.ProjectMember) error {
	memberType := member.GetType()
	var err error
	switch memberType {
	default:
		return fmt.Errorf("unhandled member type: %s", memberType)
	case "User":
		var m *projectMemberEntity
		var members []*projectMemberEntity
		members, err = p.api.getUserMembers(p.id)
		if err != nil {
			return err
		}
		for _, memb := range members {
			if memb.EntityName == member.GetName() {
				m = memb
				break
			}
		}
		if m == nil {
			return fmt.Errorf("user member not found")
		}
		err = p.api.deleteProjectUserMember(p.id, m.Id)
	case "Robot":
		var m *robot
		var members []*robot
		members, err = p.api.getRobotMembers(p.id)
		if err != nil {
			return err
		}
		expectedName := fmt.Sprintf("robot$%s+%s", p.GetName(), member.GetName())
		for _, memb := range members {
			if memb.GetName() == expectedName {
				m = memb
				break
			}
		}
		if m == nil {
			return fmt.Errorf("robot member not found")
		}
		err = p.api.deleteProjectRobotMember(p.id, m.Id)
	}
	return err
}

func (p *project) AssignReplicationRule(remoteReg globalregistry.RegistryConfig, trigger globalregistry.ReplicationTrigger, direction globalregistry.ReplicationDirection) (globalregistry.ReplicationRule, error) {
	return p.api.reg.ReplicationAPI().(*replicationAPI).create(p, remoteReg, trigger, direction)
}

func (p *project) getRepositories() ([]globalregistry.Repository, error) {
	return p.api.listProjectRepositories(p)
}

func (p *project) GetReplicationRules(
	trigger *globalregistry.ReplicationTrigger,
	direction *globalregistry.ReplicationDirection) ([]globalregistry.ReplicationRule, error) {
	p.api.reg.logger.V(1).Info("Project.GetReplicationRules invoked",
		"projectName", p.Name,
	)
	replRules, err := p.api.reg.replications.List()
	if err != nil {
		return nil, err
	}
	p.api.reg.logger.V(1).Info("replication rules fetched",
		"count", len(replRules),
	)
	results := make([]globalregistry.ReplicationRule, 0)
	for _, replRule := range replRules {
		p.api.reg.logger.V(1).Info("checking replication rule",
			"name", replRule.GetName(),
			"projectName", replRule.GetProjectName(),
		)
		if replRule.GetProjectName() == p.Name {
			p.api.reg.logger.V(1).Info("project name matches, replication rule stored")
			if trigger != nil && *trigger != replRule.Trigger() {
				continue
			}
			if direction != nil && *direction != replRule.Direction() {
				continue
			}
			results = append(results, replRule)
		}
	}
	return results, nil
}

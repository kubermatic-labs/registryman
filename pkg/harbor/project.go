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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type project struct {
	id   int
	api  *projectAPI
	Name string
	sApi *scannerAPI
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
		switch opt := p.api.reg.GetOptions().(type) {
		case globalregistry.CanForceDelete:
			if f := opt.ForceDeleteProjects(); !f {
				return fmt.Errorf("%s: repositories are present, please delete them before deleting the project, %w", p.Name, globalregistry.ErrRecoverableError)
			}
			for _, repo := range repos {
				p.api.reg.logger.V(1).Info("deleting repository",
					"repositoryName", repo.Name,
				)
				err = p.deleteRepository(repo)
				if err != nil {
					return err
				}
			}
		}

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
	case userType:
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
		return nil, err
	case groupType:
		groupMember, ok := member.(globalregistry.LdapMember)
		if !ok {
			return nil, fmt.Errorf("error assigning group %s to project %s: group is not LDAP group",
				member.GetName(), p.Name)
		}
		role, err := roleFromString(member.GetRole())
		if err != nil {
			return nil, err
		}
		userGroup := &userGroup{
			GroupName:   member.GetName(),
			LdapGroupDn: groupMember.GetDN(),
			GroupType:   1,
		}

		_, err = p.api.reg.updateIDOfUserGroup(userGroup)
		// err = p.api.reg.getOrCreateUsergroup(userGroup)
		if err != nil {
			return nil, err
		}

		pum := &projectMemberRequestBody{
			RoleId:      role,
			MemberGroup: userGroup,
		}
		_, err = p.api.createProjectMember(p.id, pum)
		return nil, err
	case robotType:
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
	userGroupMembers, err := p.api.getMembers(p.id)
	if err != nil {
		return nil, err
	}
	robotMembers, err := p.api.getRobotMembers(p.id)
	if err != nil {
		return nil, err
	}
	members := make([]globalregistry.ProjectMember, len(userGroupMembers)+len(robotMembers))

	c := 0
	// collecting the members of type User
	for _, userGroup := range userGroupMembers {
		members[c] = userGroup.toProjectMember()
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
	case userType, groupType:
		var m *projectMemberEntity
		var members []*projectMemberEntity
		members, err = p.api.getMembers(p.id)
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
		err = p.api.deleteProjectMember(p.id, m.Id)
	case robotType:
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

func (p *project) getRepositories() ([]*projectRepositoryRespBody, error) {
	return p.api.listProjectRepositories(p)
}

func (p *project) deleteRepository(r *projectRepositoryRespBody) error {
	return p.api.deleteProjectRepository(p, r)
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

func (p *project) GetScanner() (globalregistry.Scanner, error) {
	return p.sApi.getForProject(p.id)
}

func (p *project) AssignScanner(targetScanner globalregistry.Scanner) error {
	scannerID, err := p.sApi.getScannerIDByNameOrCreate(targetScanner)
	if err != nil {
		return err
	}
	return p.sApi.SetForProject(p.id, scannerID)
}

func (p *project) UnassignScanner(targetScanner globalregistry.Scanner) error {
	var defaultScanner globalregistry.Scanner
	currentScanners, err := p.sApi.List()

	if err != nil {
		return fmt.Errorf("couldn't list scanners for project, %w", err)
	}

	for _, s := range currentScanners {
		if s.(*scanner).isDefault {
			defaultScanner = s
		}
	}

	if defaultScanner.GetName() == targetScanner.GetName() {
		return nil
	}
	if defaultScanner.GetName() == "" {
		p.api.reg.logger.Error(err, "couldn't find default scanner for project", p)
		return err
	}

	return p.AssignScanner(defaultScanner)
}

type projectStatusQuotaUsed struct {
	Storage int `json:"storage"`
}
type projectStatusQuota struct {
	Used projectStatusQuotaUsed `json:"used"`
}

type projectStatusResponse struct {
	Quota projectStatusQuota `json:"quota"`
}

// GetUsedStorage implements the globalregistry.Project interface.
func (p *project) GetUsedStorage() (int, error) {
	p.api.reg.logger.V(1).Info("getting storage usage of a project",
		"projectName", p.Name,
	)
	url := *p.api.reg.parsedUrl
	url.Path = fmt.Sprintf("/api/v2.0/projects/%d/summary", p.id)
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return -1, err
	}

	req.SetBasicAuth(p.api.reg.GetUsername(), p.api.reg.GetPassword())

	resp, err := p.api.reg.do(req)
	if err != nil {
		return -1, err
	}

	defer resp.Body.Close()

	parsedResponse := &projectStatusResponse{}

	err = json.NewDecoder(resp.Body).Decode(&parsedResponse)
	if err != nil {
		p.api.reg.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		p.api.reg.logger.Info(b.String())
	}
	return parsedResponse.Quota.Used.Storage, nil
}

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

	"strconv"
	"strings"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type projectMemberEntity struct {
	EntityId   int    `json:"entity_id"`
	RoleName   string `json:"role_name"`
	EntityName string `json:"entity_name"`
	EntityType string `json:"entity_type"`
	ProjectId  int    `json:"project_id"`
	Id         int    `json:"id"`
	RoleId     role   `json:"role_id"`
}

var _ globalregistry.ProjectMember = &projectMemberEntity{}

func (pme *projectMemberEntity) GetName() string {
	return pme.EntityName
}

func (pme *projectMemberEntity) GetType() string {
	return "User"
}

func (pme *projectMemberEntity) GetRole() string {
	return pme.RoleId.String()
}

type userGroup struct {
	GroupName   string `json:"group_name"`
	LdapGroupDn string `json:"ldap_group_dn"`
	GroupType   int    `json:"group_type"`
	Id          int    `json:"id"`
}

type userEntity struct {
	Username string `json:"username"`
	UserId   int    `json:"user_id"`
}

type projectMemberRequestBody struct {
	RoleId role `json:"role_id"`
	// Only one of the MemberGroup and MemberUser parameters needs to be specified!
	MemberGroup *userGroup  `json:"member_group"`
	MemberUser  *userEntity `json:"member_user"`
}

// type projectUserMember struct {
// 	Name      string
// 	Role      role
// 	projectId int
// }

func (p *projectAPI) getUserMembers(projectID int) ([]*projectMemberEntity, error) {
	p.reg.parsedUrl.Path = fmt.Sprintf("%s/%d/members", path, projectID)
	p.reg.logger.V(1).Info("creating new request", "parsedUrl", p.reg.parsedUrl.String())
	req, err := http.NewRequest(http.MethodGet, p.reg.parsedUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	p.reg.logger.V(1).Info("sending HTTP request", "req-uri", req.RequestURI)

	req.SetBasicAuth(p.reg.GetUsername(), p.reg.GetPassword())

	resp, err := p.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	projectMembersResult := []*projectMemberEntity{}

	err = json.NewDecoder(resp.Body).Decode(&projectMembersResult)
	if err != nil {
		p.reg.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		p.reg.logger.Info(b.String())
		fmt.Printf("body: %+v\n", b.String())
	}
	return projectMembersResult, err
}

func (p *projectAPI) createProjectMember(projectID int, projectMember *projectMemberRequestBody) (int, error) {
	p.reg.parsedUrl.Path = fmt.Sprintf("%s/%d/members", path, projectID)
	reqBodyBuf := bytes.NewBuffer(nil)
	err := json.NewEncoder(reqBodyBuf).Encode(projectMember)
	if err != nil {
		return 0, err
	}
	req, err := http.NewRequest(http.MethodPost, p.reg.parsedUrl.String(), reqBodyBuf)
	if err != nil {
		return 0, err
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(p.reg.GetUsername(), p.reg.GetPassword())

	resp, err := p.reg.do(req)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()
	if resp.StatusCode == 409 {
		return 0, fmt.Errorf("project member already exists, %w", globalregistry.RecoverableError)
	}

	memberID, err := strconv.Atoi(strings.TrimPrefix(
		resp.Header.Get("Location"),
		fmt.Sprintf("%s/%d/members/", path, projectID)))
	if err != nil {
		p.reg.logger.Error(err, "cannot parse member ID from response Location header",
			"location-header", resp.Header.Get("Location"))
		return 0, err
	}

	return memberID, nil
}

func (p *projectAPI) deleteProjectUserMember(projectID int, memberId int) error {
	p.reg.parsedUrl.Path = fmt.Sprintf("%s/%d/members/%d", path, projectID, memberId)
	p.reg.logger.V(1).Info("creating new request", "parsedUrl", p.reg.parsedUrl.String())
	req, err := http.NewRequest(http.MethodDelete, p.reg.parsedUrl.String(), nil)
	if err != nil {
		return err
	}
	p.reg.logger.V(1).Info("sending HTTP request", "req-uri", req.URL)

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(p.reg.GetUsername(), p.reg.GetPassword())

	resp, err := p.reg.do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}

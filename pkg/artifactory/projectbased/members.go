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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

const userType = "User"

type projectMembers struct {
	Members []projectMember `json:"members"`
}

type projectMember struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

var _ globalregistry.ProjectMember = &projectMember{}

func (m *projectMember) GetName() string {
	return m.Name
}

func (m *projectMember) GetType() string {
	return userType
}

func (m *projectMember) GetRole() string {
	return strings.Join(m.Roles, ",")
}

func (m *projectMember) toProjectMember() globalregistry.ProjectMember {
	return (*projectMember)(m)

}

func (r *projectRegistry) getMembers(ctx context.Context, projectKey string) ([]projectMember, error) {
	url := *r.parsedUrl
	url.Path = fmt.Sprintf("%s/%s/users", projectPath, projectKey)
	r.logger.V(1).Info("creating new request", "url", url.String())
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+r.getAccessToken())
	req.Header.Add("Accept", "application/json")

	resp, err := r.do(ctx, req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	projectMembersResult := &projectMembers{}

	err = json.NewDecoder(resp.Body).Decode(&projectMembersResult)
	if err != nil {
		r.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		r.logger.Info(b.String())
		fmt.Printf("body: %+v\n", b.String())
	}
	return projectMembersResult.Members, err
}

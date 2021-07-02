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

package registry

import (
	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman.kubermatic.com/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type projectMember struct {
	*api.ProjectMember
}

var _ globalregistry.ProjectMember = &projectMember{}

func (member *projectMember) GetName() string {
	return member.ProjectMember.Name
}

func (member *projectMember) GetType() string {
	return member.ProjectMember.Type.String()

}
func (member *projectMember) GetRole() string {
	return member.ProjectMember.Role.String()
}

type ldapGroupMember struct {
	*projectMember
}

var _ globalregistry.LdapMember = &ldapGroupMember{}

func (member *ldapGroupMember) GetDN() string {
	return member.DN
}

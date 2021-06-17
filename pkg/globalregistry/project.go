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
package globalregistry

type ProjectMember interface {
	GetName() string
	GetType() string
	GetRole() string
}

// LdapMember is a ProjectMember that is stored in Ldap and as such it has a
// distinguished name (DN).
type LdapMember interface {
	ProjectMember
	GetDN() string
}

type Repository interface {
	GetName() string
	Delete() error
}

// ProjectMemberCredentials contains the username and password of a member
// (typically of type robot) that is created during the AssignMember operation
// of a Project.
type ProjectMemberCredentials struct {
	Username string
	Password string
}

type Project interface {
	AssignMember(ProjectMember) (*ProjectMemberCredentials, error)
	UnassignMember(ProjectMember) error
	GetMembers() ([]ProjectMember, error)

	AssignReplicationRule(RegistryConfig, ReplicationTrigger, ReplicationDirection) (ReplicationRule, error)
	Delete() error
	GetName() string
	GetReplicationRules(*ReplicationTrigger, *ReplicationDirection) ([]ReplicationRule, error)
}

type ProjectAPI interface {
	Create(name string) (Project, error)
	GetByName(name string) (Project, error)
	List() ([]Project, error)
}

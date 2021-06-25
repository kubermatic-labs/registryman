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

// ProjectMember interface defines the methods that are common for all types of
// project members.
type ProjectMember interface {

	// GetName method returns with the name of the project member.
	GetName() string

	// GetType method returns with the type of the project member, e.g.
	// user, group, robot, etc.
	GetType() string

	// GetRole method returns with the role of the project member, e.g.
	// Maintainer, Administrator, etc.
	GetRole() string
}

// LdapMember is a ProjectMember that is stored in Ldap and as such it has a
// distinguished name (DN).
type LdapMember interface {
	ProjectMember
	GetDN() string
}

// ProjectMemberCredentials contains the username and password of a member
// (typically of type robot) that is created during the AssignMember operation
// of a Project.
type ProjectMemberCredentials struct {
	Username string
	Password string
}

// Project interface defines the methods that can be performed on a project of a
// registry.
type Project interface {

	// GetName returns the name of the project.
	GetName() string

	// Delete removes the project from the registry.
	Delete() error

	// GetMembers returns the list of project members.
	GetMembers() ([]ProjectMember, error)

	// AssignMember method assigns a project member (user, group or robot)
	// to a project. When credentials are created by the registry provider,
	// they are returned. Otherwise, ProjectMemberCredentials is nil.
	AssignMember(ProjectMember) (*ProjectMemberCredentials, error)

	// UnassignMember removes a project member from the project.
	UnassignMember(ProjectMember) error

	// GetReplicationRules returns the list of replication rule concerning
	// the project of the registry.
	GetReplicationRules(*ReplicationTrigger, *ReplicationDirection) ([]ReplicationRule, error)

	// AssignReplicationRule assigns a replication rule to the project.
	AssignReplicationRule(RegistryConfig, ReplicationTrigger, ReplicationDirection) (ReplicationRule, error)

	// GetScanner returns the scanner assigned to the project.
	GetScanner() (Scanner, error)

	// AssignScanner assigns a scanner to the project.
	AssignScanner(Scanner) error

	// UnassignScanner removes a scanner from the project.
	UnassignScanner(Scanner) error

	Storage
}

// ProjectAPI interface defines the methods of a registry which are related to
// the management of the projects.
type ProjectAPI interface {

	// List returns the list of the projects managed by the registry.
	List() ([]Project, error)

	// GetByName returns the project with the given name. If no project is
	// present with the given name (nil, nil) is returned.
	GetByName(name string) (Project, error)

	// Create creates a new project with the given name.
	Create(name string) (Project, error)
}

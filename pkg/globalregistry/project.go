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

import "context"

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
}

// ProjectWithRepositories interface defines the methods that can be performed
// on a project of a registry that has repositories.
type ProjectWithRepositories interface {
	// GetRepositories returns the repositories found in the project
	GetRepositories(context.Context) ([]string, error)
}

// Project interface defines the methods that can be performed on a project of a
// registry which can delete the given project.
type DestructibleProject interface {
	// Delete removes the project from the registry.
	Delete(context.Context) error
}

// ProjectWithMembers interface contains the methods that we use for
// project-level member related read-only operations.
type ProjectWithMembers interface {
	// GetMembers returns the list of project members.
	GetMembers(context.Context) ([]ProjectMember, error)
}

// MemberManipulatorProject interface contains the methods that we use for
// project-level member related read-write operations.
type MemberManipulatorProject interface {
	// AssignMember method assigns a project member (user, group or robot)
	// to a project. When credentials are created by the registry provider,
	// they are returned. Otherwise, ProjectMemberCredentials is nil.
	AssignMember(context.Context, ProjectMember) (*ProjectMemberCredentials, error)

	// UnassignMember removes a project member from the project.
	UnassignMember(context.Context, ProjectMember) error
}

// ProjectWithScanner interface contains the methods that we use for
// project-level scanner related read-only operations.
type ProjectWithScanner interface {
	// GetScanner returns the scanner assigned to the project.
	GetScanner(context.Context) (Scanner, error)
}

// ScannerManipulatorProject interface contains the methods that we use for
// project-level scanner related read-write operations.
type ScannerManipulatorProject interface {
	// AssignScanner assigns a scanner to the project.
	AssignScanner(context.Context, Scanner) error

	// UnassignScanner removes a scanner from the project.
	UnassignScanner(context.Context, Scanner) error
}

// ProjectWithReplication interface contains the methods that we use for
// project-level replication related read-only operations.
type ProjectWithReplication interface {
	// GetReplicationRules returns the list of replication rule concerning
	// the project of the registry.
	GetReplicationRules(ctx context.Context, trigger, direction string) ([]ReplicationRule, error)
}

// ReplicationRuleManipulatorProject interface contains the methods that we use
// for project-level replication related read-write manipulations.
type ReplicationRuleManipulatorProject interface {
	// AssignReplicationRule assigns a replication rule to the project.
	AssignReplicationRule(ctx context.Context, remote Registry, trigger, direction string) (ReplicationRule, error)
}

// ProjectWithStorage interface contains the methods that we use for
// project-level storage related operations.
type ProjectWithStorage interface {
	// GetUsedStorage returns the used storage in bytes.
	GetUsedStorage(context.Context) (int, error)
}

// RegistryWithProjects interface defines the methods of a registry which are
// related to the management of the projects.
type RegistryWithProjects interface {
	// List returns the list of the projects managed by the registry.
	ListProjects(context.Context) ([]Project, error)

	// GetByName returns the project with the given name. If no project is
	// present with the given name (nil, nil) is returned.
	//
	// For empty project name a dummy Project is created that can be used
	// for testing the registry's capabilities.
	GetProjectByName(ctx context.Context, name string) (Project, error)
}

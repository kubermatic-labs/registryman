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
package reconciler

import (
	"context"

	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type nilSideEffect struct{}

func (se nilSideEffect) Perform(context.Context) error {
	return nil
}

var nilEffect = nilSideEffect{}

type SideEffectContextKey string

var SideEffectPath = SideEffectContextKey("path")
var SideEffectSerializer = SideEffectContextKey("serializer")

type SideEffect interface {
	Perform(ctx context.Context) error
}

type Action interface {
	String() string
	Perform(globalregistry.Registry) (SideEffect, error)
}

type RegistryStatus struct {
	Projects []ProjectStatus
}

func Compare(store *config.ExpectedProvider, actual, expected *RegistryStatus) []Action {
	return CompareProjectStatuses(store, actual.Projects, expected.Projects)
}

func GetRegistryStatus(reg globalregistry.Registry) (*RegistryStatus, error) {
	projects, err := reg.ProjectAPI().List()
	if err != nil {
		return nil, err
	}
	projectStatuses := make([]ProjectStatus, len(projects))
	for i, project := range projects {
		projectStatuses[i].Name = project.GetName()

		members, err := project.GetMembers()
		if err != nil {
			return nil, err
		}
		projectStatuses[i].Members = make([]MemberStatus, len(members))
		for n, member := range members {
			projectStatuses[i].Members[n].Name = member.GetName()
			projectStatuses[i].Members[n].Type = member.GetType()
			projectStatuses[i].Members[n].Role = member.GetRole()
			switch m := member.(type) {
			case globalregistry.LdapMember:
				projectStatuses[i].Members[n].DN = m.GetDN()
			}
		}
		replicationRules, err := project.GetReplicationRules(nil, nil)
		if err != nil {
			return nil, err
		}
		projectStatuses[i].ReplicationRules = make([]ReplicationRuleStatus, len(replicationRules))
		for n, rule := range replicationRules {
			projectStatuses[i].ReplicationRules[n].RemoteRegistryName = rule.RemoteRegistry().GetName()
			projectStatuses[i].ReplicationRules[n].Trigger = rule.Trigger()
			projectStatuses[i].ReplicationRules[n].Direction = rule.Direction()
		}
	}
	return &RegistryStatus{
		Projects: projectStatuses,
	}, nil
}

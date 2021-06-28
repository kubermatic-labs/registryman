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
	"errors"

	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

// RegistryStatus specifies the status of a registry.
type RegistryStatus struct {
	Projects []ProjectStatus `json:"projects"`
}

// Compare compares the actual and expected status of a registry. The function
// returns the actions that are needed to synchronize the actual state to the
// expected state.
func Compare(store *config.ExpectedProvider, actual, expected *RegistryStatus) []Action {
	return CompareProjectStatuses(store, actual.Projects, expected.Projects)
}

// GetRegistryStatus function calculate the status of a registry. If the
// registry represents a configuration of registry, then the expected registry
// status is returned. If the registry represents an actual (real) registry, the
// actual status is returned.
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

		storageUsed, err := project.GetUsedStorage()
		switch {
		case errors.Is(err, globalregistry.ErrNotImplemented):
			// we use the default value, if GetUsedStorage is not implemented
		case err == nil:
			projectStatuses[i].StorageUsed = storageUsed
		default:
			return nil, err
		}

		projectScanner, err := project.GetScanner()
		if err != nil {
			return nil, err
		}
		if projectScanner != nil {
			projectStatuses[i].ScannerStatus = ScannerStatus{
				Name: projectScanner.GetName(),
				URL:  projectScanner.GetURL(),
			}
		}
	}
	return &RegistryStatus{
		Projects: projectStatuses,
	}, nil
}

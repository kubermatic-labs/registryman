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

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

// Compare compares the actual and expected status of a registry. The function
// returns the actions that are needed to synchronize the actual state to the
// expected state.
func Compare(store *config.ExpectedProvider, actual, expected *api.RegistryStatus) []Action {
	return CompareProjectStatuses(store, actual.Projects, expected.Projects, actual.Capabilities)
}

func getRegistryCapabilities(ctx context.Context, reg globalregistry.Registry) (api.RegistryCapabilities, error) {
	replCap := globalregistry.GetReplicationCapability(reg.GetProvider())
	regWithProjects := reg.(globalregistry.RegistryWithProjects)
	registryCapabilities := api.RegistryCapabilities{
		CanPullReplicate: replCap.CanPull(),
		CanPushReplicate: replCap.CanPush(),
	}
	dummyProject, err := regWithProjects.GetProjectByName(ctx, "")
	if err != nil {
		return registryCapabilities, err
	}

	if _, ok := reg.(globalregistry.ProjectCreator); ok {
		registryCapabilities.CanCreateProject = true
	}
	if _, ok := dummyProject.(globalregistry.DestructibleProject); ok {
		registryCapabilities.CanDeleteProject = true
	}
	if _, ok := dummyProject.(globalregistry.ProjectWithMembers); ok {
		registryCapabilities.HasProjectMembers = true
	}
	if _, ok := dummyProject.(globalregistry.MemberManipulatorProject); ok {
		registryCapabilities.CanManipulateProjectMembers = true
	}
	if _, ok := dummyProject.(globalregistry.ProjectWithScanner); ok {
		registryCapabilities.HasProjectScanners = true
	}
	if _, ok := dummyProject.(globalregistry.ScannerManipulatorProject); ok {
		registryCapabilities.CanManipulateProjectScanners = true
	}
	if _, ok := dummyProject.(globalregistry.ProjectWithReplication); ok {
		registryCapabilities.HasProjectReplicationRules = true
	}
	if _, ok := dummyProject.(globalregistry.ReplicationRuleManipulatorProject); ok {
		registryCapabilities.CanManipulateProjectReplicationRules = true
	}
	if _, ok := dummyProject.(globalregistry.ProjectWithStorage); ok {
		registryCapabilities.HasProjectStorageReport = true
	}
	return registryCapabilities, nil
}

// GetRegistryStatus function calculate the status of a registry. If the
// registry represents a configuration of registry, then the expected registry
// status is returned. If the registry represents an actual (real) registry, the
// actual status is returned.
func GetRegistryStatus(ctx context.Context, reg globalregistry.Registry) (*api.RegistryStatus, error) {
	regWithProjects := reg.(globalregistry.RegistryWithProjects)
	registryCapabilities, err := getRegistryCapabilities(ctx, reg)
	if err != nil {
		return nil, err
	}
	projects, err := regWithProjects.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	projectStatuses := make([]api.ProjectStatus, len(projects))
	for i, project := range projects {
		projectStatuses[i].Name = project.GetName()

		projectWithMembers, ok := project.(globalregistry.ProjectWithMembers)
		if ok {
			// registry supports projects with members
			members, err := projectWithMembers.GetMembers(ctx)
			if err != nil {
				return nil, err
			}
			projectStatuses[i].Members = make([]api.MemberStatus, len(members))
			for n, member := range members {
				projectStatuses[i].Members[n].Name = member.GetName()
				projectStatuses[i].Members[n].Type = member.GetType()
				projectStatuses[i].Members[n].Role = member.GetRole()
				switch m := member.(type) {
				case globalregistry.LdapMember:
					projectStatuses[i].Members[n].DN = m.GetDN()
				}
			}
		} else {
			projectStatuses[i].Members = make([]api.MemberStatus, 0)
		}
		projectWithReplication, ok := project.(globalregistry.ProjectWithReplication)
		if ok {
			replicationRules, err := projectWithReplication.GetReplicationRules(ctx, "", "")
			if err != nil {
				return nil, err
			}
			projectStatuses[i].ReplicationRules = make([]api.ReplicationRuleStatus, len(replicationRules))
			for n, rule := range replicationRules {
				projectStatuses[i].ReplicationRules[n].RemoteRegistryName = rule.RemoteRegistry().GetName()
				projectStatuses[i].ReplicationRules[n].Trigger = string(rule.Trigger())
				projectStatuses[i].ReplicationRules[n].Direction = rule.Direction()
			}
		} else {
			projectStatuses[i].ReplicationRules = make([]api.ReplicationRuleStatus, 0)
		}

		projectWithStorage, ok := project.(globalregistry.ProjectWithStorage)
		if ok {
			storageUsed, err := projectWithStorage.GetUsedStorage(ctx)
			if err != nil {
				return nil, err
			}
			projectStatuses[i].StorageUsed = storageUsed
		}

		projectWithScanner, ok := project.(globalregistry.ProjectWithScanner)
		if ok {
			projectScanner, err := projectWithScanner.GetScanner(ctx)
			if err != nil {
				return nil, err
			}
			if projectScanner != nil {
				projectStatuses[i].ScannerStatus = api.ScannerStatus{
					Name: projectScanner.GetName(),
					URL:  projectScanner.GetURL(),
				}
			}
		}
	}
	return &api.RegistryStatus{
		Projects:     projectStatuses,
		Capabilities: registryCapabilities,
	}, nil
}

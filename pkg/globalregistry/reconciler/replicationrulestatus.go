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
	"fmt"

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman.kubermatic.com/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type rRuleAddAction struct {
	api.ReplicationRuleStatus
	store       *config.ExpectedProvider
	projectName string
}

var _ Action = &rRuleAddAction{}

func (ra *rRuleAddAction) String() string {
	return fmt.Sprintf("adding replication rule for %s: %s [%s] on %s",
		ra.projectName,
		ra.RemoteRegistryName,
		ra.Direction,
		ra.Trigger,
	)
}

func (ra *rRuleAddAction) Perform(reg globalregistry.Registry) (SideEffect, error) {
	project, err := reg.(globalregistry.RegistryWithProjects).GetProjectByName(ra.projectName)
	if err != nil {
		return nilEffect, err
	}
	remoteRegistry := ra.store.GetRegistryByName(ra.RemoteRegistryName)
	if remoteRegistry == nil {
		return nilEffect, fmt.Errorf("registry %s not found in object store", ra.RemoteRegistryName)
	}
	replicationRuleManipulatorProject, ok := project.(globalregistry.ReplicationRuleManipulatorProject)
	if !ok {
		// registry does not support project level replication
		return nilEffect, nil
	}
	_, err = replicationRuleManipulatorProject.AssignReplicationRule(remoteRegistry, ra.Trigger, ra.Direction)
	return nilEffect, err
}

type rRuleRemoveAction struct {
	api.ReplicationRuleStatus
	store       *config.ExpectedProvider
	projectName string
}

var _ Action = &rRuleRemoveAction{}

func (ra *rRuleRemoveAction) String() string {
	return fmt.Sprintf("removing replication rule for %s: %s [%s] on %s",
		ra.projectName,
		ra.RemoteRegistryName,
		ra.Direction,
		ra.Trigger,
	)
}

func (ra *rRuleRemoveAction) Perform(reg globalregistry.Registry) (SideEffect, error) {
	project, err := reg.(globalregistry.RegistryWithProjects).GetProjectByName(ra.projectName)
	if err != nil {
		return nilEffect, err
	}
	projectWithReplication, ok := project.(globalregistry.ProjectWithReplication)
	if !ok {
		// registry does not support project level replication
		return nilEffect, nil
	}
	rRules, err := projectWithReplication.GetReplicationRules(ra.Trigger, ra.Direction)
	if err != nil {
		return nilEffect, err
	}
	for _, rRule := range rRules {
		destructibleReplicationRule, ok := rRule.(globalregistry.DestructibleReplicationRule)
		if !ok {
			// TODO: error handling
			continue
		}
		err := destructibleReplicationRule.Delete()
		if err != nil {
			return nilEffect, err
		}
	}
	return nilEffect, nil
}

// CompareReplicationRuleStatus compares the actual and expected status of the
// replication rules of a project. The function returns the actions that are
// needed to synchronize the actual state to the expected state.
func CompareReplicationRuleStatus(store *config.ExpectedProvider, projectName string, actual, expected []api.ReplicationRuleStatus) []Action {
	actualDiff := []api.ReplicationRuleStatus{}
	expectedDiff := []api.ReplicationRuleStatus{}
ActLoop:
	for _, act := range actual {
		for _, exp := range expected {
			if act == exp {
				continue ActLoop
			}
		}
		// act was not found among expected rules
		actualDiff = append(actualDiff, act)
	}
ExpLoop:
	for _, exp := range expected {
		for _, act := range actual {
			if act == exp {
				continue ExpLoop
			}
		}
		// exp was not found among actual rules
		expectedDiff = append(expectedDiff, exp)
	}
	actions := make([]Action, 0)

	// actualDiff contains the members which are there but are not needed
	for _, act := range actualDiff {
		actions = append(actions, &rRuleRemoveAction{
			act,
			store,
			projectName,
		})
	}

	// expectedClone contains the members which are missing and thus they
	// shall be created
	for _, exp := range expectedDiff {
		actions = append(actions, &rRuleAddAction{
			exp,
			store,
			projectName,
		})
	}

	return actions
}

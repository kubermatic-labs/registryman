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

	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

// ReplicationRuleStatus specifies the status of project replication rule.
type ReplicationRuleStatus struct {

	// RemoteRegistryName indicates the name of the remote registry which
	// the current registry shall synchronize with.
	RemoteRegistryName string `json:"name"`

	// Trigger describes the event that shall trigger the replication.
	Trigger globalregistry.ReplicationTrigger `json:"trigger"`

	// Direction shows whether the replication is of type pull or push.
	Direction globalregistry.ReplicationDirection `json:"direction"`
}

type rRuleAddAction struct {
	ReplicationRuleStatus
	store       *config.ExpectedProvider
	projectName string
}

var _ Action = &rRuleAddAction{}

func (ra *rRuleAddAction) String() string {
	return fmt.Sprintf("adding replication rule for %s: %s [%s] on %s",
		ra.projectName,
		ra.RemoteRegistryName,
		ra.Direction.String(),
		ra.Trigger.String(),
	)
}

func (ra *rRuleAddAction) Perform(reg globalregistry.Registry) (SideEffect, error) {
	project, err := reg.ProjectAPI().GetByName(ra.projectName)
	if err != nil {
		return nilEffect, err
	}
	remoteRegistry := ra.store.GetRegistryByName(ra.RemoteRegistryName)
	if remoteRegistry == nil {
		return nilEffect, fmt.Errorf("registry %s not found in object store", ra.RemoteRegistryName)
	}
	_, err = project.AssignReplicationRule(remoteRegistry, ra.Trigger, ra.Direction)
	return nilEffect, err
}

type rRuleRemoveAction struct {
	ReplicationRuleStatus
	store       *config.ExpectedProvider
	projectName string
}

var _ Action = &rRuleRemoveAction{}

func (ra *rRuleRemoveAction) String() string {
	return fmt.Sprintf("removing replication rule for %s: %s [%s] on %s",
		ra.projectName,
		ra.RemoteRegistryName,
		ra.Direction.String(),
		ra.Trigger.String(),
	)
}

func (ra *rRuleRemoveAction) Perform(reg globalregistry.Registry) (SideEffect, error) {
	project, err := reg.ProjectAPI().GetByName(ra.projectName)
	if err != nil {
		return nilEffect, err
	}
	rRules, err := project.GetReplicationRules(&ra.Trigger, &ra.Direction)
	if err != nil {
		return nilEffect, err
	}
	for _, rRule := range rRules {
		err := rRule.Delete()
		if err != nil {
			return nilEffect, err
		}
	}
	return nilEffect, nil
}

// CompareReplicationRuleStatus compares the actual and expected status of the
// replication rules of a project. The function returns the actions that are
// needed to synchronize the actual state to the expected state.
func CompareReplicationRuleStatus(store *config.ExpectedProvider, projectName string, actual, expected []ReplicationRuleStatus) []Action {
	actualDiff := []ReplicationRuleStatus{}
	expectedDiff := []ReplicationRuleStatus{}
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

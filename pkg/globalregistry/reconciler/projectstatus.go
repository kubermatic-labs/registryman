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

type ProjectStatus struct {
	Name             string
	Members          []MemberStatus
	ReplicationRules []ReplicationRuleStatus
}

type projectAddAction struct {
	ProjectStatus
}

var _ Action = &projectAddAction{}

func (pa *projectAddAction) String() string {
	return fmt.Sprintf("adding project %s", pa.Name)
}

func (pa *projectAddAction) Perform(reg globalregistry.Registry) (SideEffect, error) {
	_, err := reg.ProjectAPI().Create(pa.Name)
	if err != nil {
		return nilEffect, err
	}
	return nilEffect, nil
}

type projectRemoveAction struct {
	ProjectStatus
}

var _ Action = &projectRemoveAction{}

func (pa *projectRemoveAction) String() string {
	return fmt.Sprintf("removing project %s", pa.Name)
}

func (pa *projectRemoveAction) Perform(reg globalregistry.Registry) (SideEffect, error) {
	project, err := reg.ProjectAPI().GetByName(pa.Name)
	if err != nil {
		return nilEffect, err
	}
	return nilEffect, project.Delete()
}

func CompareProjectStatuses(store *config.ExpectedProvider, actual, expected []ProjectStatus) []Action {
	same := make(map[string][2]ProjectStatus)
	actualDiff := []ProjectStatus{}
	expectedDiff := []ProjectStatus{}
ActLoop:
	for _, act := range actual {
		for _, exp := range expected {
			if act.Name == exp.Name {
				same[act.Name] = [2]ProjectStatus{
					act,
					exp,
				}
				continue ActLoop
			}
		}
		// act was not found among expected members
		actualDiff = append(actualDiff, act)
	}
ExpLoop:
	for _, exp := range expected {
		for _, act := range actual {
			if act.Name == exp.Name {
				// we have already found this in the ActLoop
				continue ExpLoop
			}
		}
		// exp was not found among actual members
		expectedDiff = append(expectedDiff, exp)
	}
	actions := make([]Action, 0)

	// actualDiff contains the projects which are there but are not needed
	for _, act := range actualDiff {
		// We remove the related replication rules first
		for _, replRule := range act.ReplicationRules {
			actions = append(actions, &rRuleRemoveAction{
				ReplicationRuleStatus: replRule,
				store:                 store,
				projectName:           act.Name,
			})

		}
		// Then remove the project itself
		actions = append(actions, &projectRemoveAction{
			act,
		})
	}

	// same contains the projects that are present in both actual and
	// expected. They have to be checked for member and replication rule
	// differences.
	for projectName, projectPair := range same {
		actions = append(actions,
			CompareMemberStatuses(projectName,
				projectPair[0].Members,
				projectPair[1].Members)...,
		)
		actions = append(actions,
			CompareReplicationRuleStatus(store,
				projectName,
				projectPair[0].ReplicationRules,
				projectPair[1].ReplicationRules)...,
		)
	}
	// expectedClone contains the projects which are missing and thus they
	// shall be created
	for _, exp := range expectedDiff {
		actions = append(actions, &projectAddAction{
			exp,
		})
		for _, member := range exp.Members {
			actions = append(actions, &memberAddAction{
				member,
				exp.Name,
			})
		}
		for _, replRule := range exp.ReplicationRules {
			actions = append(actions, &rRuleAddAction{
				ReplicationRuleStatus: replRule,
				store:                 store,
				projectName:           exp.Name,
			})

		}
	}

	return actions
}

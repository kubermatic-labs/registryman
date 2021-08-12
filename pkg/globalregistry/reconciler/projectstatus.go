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

type projectAddAction struct {
	api.ProjectStatus
}

var _ Action = &projectAddAction{}

func (pa *projectAddAction) String() string {
	return fmt.Sprintf("adding project %s", pa.Name)
}

func (pa *projectAddAction) Perform(reg globalregistry.Registry) (SideEffect, error) {
	papi, ok := reg.(globalregistry.ProjectCreater)
	if !ok {
		// registry provider does not implement project creation
		return nilEffect, nil
	}
	_, err := papi.CreateProject(pa.Name)
	return nilEffect, err
}

type projectRemoveAction struct {
	api.ProjectStatus
}

var _ Action = &projectRemoveAction{}

func (pa *projectRemoveAction) String() string {
	return fmt.Sprintf("removing project %s", pa.Name)
}

func (pa *projectRemoveAction) Perform(reg globalregistry.Registry) (SideEffect, error) {
	project, err := reg.(globalregistry.RegistryWithProjects).GetProjectByName(pa.Name)
	if err != nil {
		return nilEffect, err
	}
	destructibleProject, ok := project.(globalregistry.DestructibleProject)
	if !ok {
		return nilEffect, nil
	}
	return nilEffect, destructibleProject.Delete()
}

// CompareProjectStatuses compares the actual and expected status of the projects
// of a registry. The function returns the actions that are needed to synchronize
// the actual state to the expected state.
func CompareProjectStatuses(store *config.ExpectedProvider, actual, expected []api.ProjectStatus) []Action {
	same := make(map[string][2]api.ProjectStatus)
	actualDiff := []api.ProjectStatus{}
	expectedDiff := []api.ProjectStatus{}
ActLoop:
	for _, act := range actual {
		for _, exp := range expected {
			if act.Name == exp.Name {
				same[act.Name] = [2]api.ProjectStatus{
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
			ProjectStatus: act,
		})
	}

	// same contains the projects that are present in both actual and
	// expected. They have to be checked for member, replication and scanner rule
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
		actions = append(actions,
			CompareScannerStatuses(
				projectName,
				projectPair[0].ScannerStatus,
				projectPair[1].ScannerStatus)...,
		)
	}
	// expectedDiff contains the projects which are missing and thus they
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
		if exp.ScannerStatus.Name != "" {
			actions = append(actions, &scannerAssignAction{
				projectName:   exp.Name,
				ScannerStatus: &exp.ScannerStatus,
			})
		}
	}

	return actions
}

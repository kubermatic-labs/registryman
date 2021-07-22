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
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type scannerStatus api.ScannerStatus

var _ globalregistry.Scanner = &scannerStatus{}

// GetName implements the globalregistry.Scanner interface.
func (ss *scannerStatus) GetName() string {
	return ss.Name
}

// GetURL implements the globalregistry.Scanner interface.
func (ss *scannerStatus) GetURL() string {
	return ss.URL
}

type scannerAssignAction struct {
	projectName string
	*api.ScannerStatus
}

var _ Action = &scannerAssignAction{}

func (a *scannerAssignAction) String() string {
	return fmt.Sprintf("assigning scanner %s to project %s",
		a.Name, a.projectName)
}

func (a *scannerAssignAction) Perform(reg globalregistry.Registry) (SideEffect, error) {
	project, err := reg.ProjectAPI().GetByName(a.projectName)
	if err != nil {
		return nilEffect, err
	}
	err = project.AssignScanner((*scannerStatus)(a.ScannerStatus))
	return nilEffect, err
}

type scannerUnassignAction struct {
	projectName string
	*api.ScannerStatus
}

func (a *scannerUnassignAction) String() string {
	return fmt.Sprintf("unassigning scanner %s from project %s",
		a.Name, a.projectName)
}

func (a *scannerUnassignAction) Perform(reg globalregistry.Registry) (SideEffect, error) {
	project, err := reg.ProjectAPI().GetByName(a.projectName)
	if err != nil {
		return nilEffect, err
	}
	err = project.UnassignScanner((*scannerStatus)(a.ScannerStatus))
	return nilEffect, err
}

var _ Action = &scannerUnassignAction{}

// CompareScannerStatuses compares the actual and expected status of the scanner
// of a project. The function returns the actions that are needed to synchronize
// the actual state to the expected state.
func CompareScannerStatuses(projectName string, actual, expected api.ScannerStatus) []Action {
	actions := make([]Action, 0)

	// Old scanner shall be deleted
	if actual.Name != "" && actual != expected {
		actions = append(actions, &scannerUnassignAction{
			projectName:   projectName,
			ScannerStatus: &actual,
		})
	}

	// New scanner shall be configured
	if expected.Name != "" && actual != expected {
		actions = append(actions, &scannerAssignAction{
			projectName:   projectName,
			ScannerStatus: &expected,
		})
	}
	return actions
}

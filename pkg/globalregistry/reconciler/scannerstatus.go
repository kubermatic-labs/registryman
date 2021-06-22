package reconciler

import (
	"fmt"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type ScannerStatus struct {
	Name string
	Url  string
}

var _ globalregistry.Scanner = &ScannerStatus{}

func (ss *ScannerStatus) GetName() string {
	return ss.Name
}

func (ss *ScannerStatus) GetURL() string {
	return ss.Url
}

type scannerAssignAction struct {
	projectName string
	*ScannerStatus
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
	err = project.AssignScanner(a)
	return nilEffect, err
}

type scannerUnassignAction struct {
	projectName string
	*ScannerStatus
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
	err = project.UnassignScanner(a)
	return nilEffect, err
}

var _ Action = &scannerUnassignAction{}

func CompareScannerStatuses(projectName string, actual, expected ScannerStatus) []Action {
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

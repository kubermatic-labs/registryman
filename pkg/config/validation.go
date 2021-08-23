package config

import (
	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
)

// ValidateConsistent performs all validations that require the full context,
// i.e. all resources parsed. If there is a consistency issue, an error is
// returned.
func ValidateConsistency(aos ApiObjectStore) error {
	registries := aos.GetRegistries()
	projects := aos.GetProjects()
	scanners := aos.GetScanners()

	// Forcing maximum one Global registry
	err := checkGlobalRegistryCount(registries)
	if err != nil {
		return err
	}

	// Checking local registry names in all local projects
	err = checkLocalRegistryNamesInProjects(registries, projects)
	if err != nil {
		return err
	}

	// Checking scanner names in all projects
	err = checkScannerNamesInProjects(projects, scanners)
	if err != nil {
		return err
	}

	// Checking scanner name uniqueness
	err = checkScannerNameUniqueness(scanners)
	if err != nil {
		return err
	}

	// Checking project name uniqueness
	err = checkProjectNameUniqueness(projects)
	if err != nil {
		return err
	}

	// Checking registry name uniqueness
	err = checkRegistryNameUniqueness(registries)
	if err != nil {
		return err
	}
	return nil
}

// checkProject checks the validation rules of the project resources. This
// function contains the checks that can be performed on a single Project
// resource.
func checkProject(project *api.Project) error {
	var err error
	for _, member := range project.Spec.Members {
		if member.Type == api.GroupMemberType && member.DN == "" {
			logger.V(-1).Info("project has group member without DN",
				"project", project.GetName(),
				"member", member.Name,
			)
			err = ErrValidationGroupWithoutDN
		}
	}
	return err
}

// checkGlobalRegistryCount checks that there is 1 or 0 registry configured with
// the type GlobalHub.
func checkGlobalRegistryCount(registries []*api.Registry) error {
	globalRegistries := make([]string, 0)
	for _, registry := range registries {
		if registry.Spec.Role == "GlobalHub" {
			globalRegistries = append(globalRegistries, registry.Name)
		}
	}
	if len(globalRegistries) >= 2 {
		for _, registry := range globalRegistries {
			logger.V(-1).Info("Multiple Global Registries found",
				"registry_name", registry)
		}
		return ErrValidationMultipleGlobalRegistries
	}
	return nil
}

// checkLocalRegistryNamesInProjects checks that the registries referenced by
// the local projects exist.
func checkLocalRegistryNamesInProjects(registries []*api.Registry, projects []*api.Project) error {
	var err error
	localRegistries := make([]string, 0)
	for _, registry := range registries {
		if registry.Spec.Role == "Local" {
			localRegistries = append(localRegistries, registry.Name)
		}
	}
	for _, project := range projects {
		if project.Spec.Type == api.LocalProjectType {
			localRegistryExists := false
			for _, localRegistry := range project.Spec.LocalRegistries {
				for _, registry := range localRegistries {
					if localRegistry == registry {
						localRegistryExists = true
						break
					}
				}
				if !localRegistryExists {
					logger.V(-1).Info("Local registry does not exist",
						"project_name", project.Name,
						"registry_name", localRegistry)
					err = ErrValidationInvalidLocalRegistryInProject
				}
				localRegistryExists = false
			}
		}
	}

	return err
}

// checkScannerNamesInProjects checks that the scanners referenced by the
// projects exist.
func checkScannerNamesInProjects(projects []*api.Project, scanners []*api.Scanner) error {
	var err error
	scannerNames := map[string]*api.Scanner{}
	for _, scanner := range scanners {
		scannerNames[scanner.GetName()] = scanner
	}
	for _, project := range projects {
		if project.Spec.Scanner != "" &&
			scannerNames[project.Spec.Scanner] == nil {
			// there is a project with invalid scanner name
			logger.V(-1).Info("Project refers to non-existing scanner",
				"project_name", project.Name,
				"scanner_name", project.Spec.Scanner)
			err = ErrValidationScannerNameReference
		}
	}

	return err
}

// checkScannerNameUniqueness checks that there are no 2 scanners with the same
// name.
func checkScannerNameUniqueness(scanners []*api.Scanner) error {
	var err error
	scannerNames := map[string]*api.Scanner{}
	for _, scanner := range scanners {
		scannerName := scanner.GetName()
		if scannerNames[scannerName] != nil {
			logger.V(-1).Info("Multiple scanners configured with the same name",
				"scanner_name", scannerName,
			)
			err = ErrValidationScannerNameNotUnique
		}
		scannerNames[scannerName] = scanner
	}
	return err
}

// checkProjectNameUniqueness checks that there are no 2 projects with the same
// name.
func checkProjectNameUniqueness(projects []*api.Project) error {
	var err error
	projectNames := map[string]*api.Project{}
	for _, project := range projects {
		projectName := project.GetName()
		if projectNames[projectName] != nil {
			logger.V(-1).Info("Multiple projects configured with the same name",
				"project_name", projectName,
			)
			err = ErrValidationProjectNameNotUnique
		}
		projectNames[projectName] = project
	}
	return err
}

// checkRegistryNameUniqueness checks that there are no 2 registries with the
// same name.
func checkRegistryNameUniqueness(registries []*api.Registry) error {
	var err error
	registryNames := map[string]*api.Registry{}
	for _, registry := range registries {
		registryName := registry.GetName()
		if registryNames[registryName] != nil {
			logger.V(-1).Info("Multiple registries configured with the same name",
				"registry_name", registryName,
			)
			err = ErrValidationRegistryNameNotUnique
		}
		registryNames[registryName] = registry
	}
	return err
}

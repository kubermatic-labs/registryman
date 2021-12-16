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

package config

import (
	"context"

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
)

// ValidateConsistency performs all validations that require the full context,
// i.e. all resources parsed. If there is a consistency issue, an error is
// returned.
func ValidateConsistency(aos ApiObjectStore) error {
	logger.V(1).Info("ValidateConsistency invoked")
	ctx := context.Background()
	registries := aos.GetRegistries(ctx)
	projects := aos.GetProjects(ctx)
	scanners := aos.GetScanners(ctx)

	// Forcing maximum one Global registry
	err := checkGlobalRegistryCount(registries)
	if err != nil {
		return err
	}

	// Checking Artifactory annotations
	err = checkArtifactoryAnnotations(registries)
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

// checkGlobalRegistryCount checks that there is 1 or 0 registry configured with
// the type GlobalHub.
func checkArtifactoryAnnotations(registries []*api.Registry) error {
	var err error
	for _, registry := range registries {
		if registry.Spec.Provider == "artifactory" {
			hasDockerRegistryNameAnnotation := false
			hasAccesTokenAnnotation := false
			for annotation := range registry.Annotations {
				if annotation == "registryman.kubermatic.com/dockerRegistryName" {
					hasDockerRegistryNameAnnotation = true
				}
				if annotation == "registryman.kubermatic.com/accessToken" {
					hasAccesTokenAnnotation = true
				}
			}
			switch {
			case hasDockerRegistryNameAnnotation && hasAccesTokenAnnotation:
				logger.V(-1).Info("Conflicting annotations found",
					"registry_name", registry)
				err = ErrValidationArtifactoryAnnotations
			case !hasDockerRegistryNameAnnotation && !hasAccesTokenAnnotation:
				logger.V(-1).Info("No Artifactory annotations found",
					"registry_name", registry)
				err = ErrValidationArtifactoryAnnotations
			}

		}

	}
	return err
}

// checkLocalRegistryNamesInProjects checks that the registries referenced by
// the local projects exist.
func checkLocalRegistryNamesInProjects(registries []*api.Registry, projects []*api.Project) error {
	logger.V(1).Info("checkLocalRegistryNamesInProjects invoked")
	var err error
	localRegistries := make([]string, 0)
	for _, registry := range registries {
		if registry.Spec.Role == "Local" {
			logger.V(1).Info("local registry found",
				"registry", registry.Name)
			localRegistries = append(localRegistries, registry.Name)
		}
	}
	for _, project := range projects {
		logger.V(1).Info("checking project",
			"project", project.Name)
		if project.Spec.Type == api.LocalProjectType {
			logger.V(1).Info("project is of type local",
				"project", project.Name)
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
	scannerNames := map[string]bool{}
	for _, scanner := range scanners {
		scannerName := scanner.GetName()
		if scannerNames[scannerName] {
			logger.V(-1).Info("Multiple scanners configured with the same name",
				"scanner_name", scannerName,
			)
			err = ErrValidationScannerNameNotUnique
		}
		scannerNames[scannerName] = true
	}
	return err
}

// checkProjectNameUniqueness checks that there are no 2 projects with the same
// name.
func checkProjectNameUniqueness(projects []*api.Project) error {
	var err error
	projectNames := map[string]bool{}
	for _, project := range projects {
		projectName := project.GetName()
		if projectNames[projectName] {
			logger.V(-1).Info("Multiple projects configured with the same name",
				"project_name", projectName,
			)
			err = ErrValidationProjectNameNotUnique
		}
		projectNames[projectName] = true
	}
	return err
}

// checkRegistryNameUniqueness checks that there are no 2 registries with the
// same name.
func checkRegistryNameUniqueness(registries []*api.Registry) error {
	var err error
	registryNames := map[string]bool{}
	for _, registry := range registries {
		registryName := registry.GetName()
		if registryNames[registryName] {
			logger.V(-1).Info("Multiple registries configured with the same name",
				"registry_name", registryName,
			)
			err = ErrValidationRegistryNameNotUnique
		}
		registryNames[registryName] = true
	}
	return err
}

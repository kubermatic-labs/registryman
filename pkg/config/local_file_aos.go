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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/kube-openapi/pkg/validation/validate"
)

// locaFileApiObjectStore is the database of the configured resources (Projects,
// Registries and Scanners) that are stored in the local filesystem.
type localFileApiObjectStore struct {
	store      map[schema.GroupVersionKind][]runtime.Object
	serializer *json.Serializer
	options    globalregistry.RegistryOptions
	path       string
}

var _ ApiObjectStore = &localFileApiObjectStore{}

func getFileName(obj runtime.Object) string {
	metaV1Object := obj.(metav1.Object)
	return fmt.Sprintf("%s.yaml", metaV1Object.GetName())
}

// WriteResource serializes the object specified by the obj parameter. The
// filename is generated from the object name by appending .yaml to it. The path
// where the file is created is set when the ReadLocalManifests function creates the
// ApiObjectStore.
func (aos *localFileApiObjectStore) WriteResource(_ context.Context, obj runtime.Object) error {
	f, err := os.Create(getFileName(obj))
	if err != nil {
		return err
	}
	defer f.Close()

	err = aos.serializer.Encode(obj, f)
	if err != nil {
		return err
	}
	return nil
}

// RemoveResource removes a file from the filesystem. The filename is generated
// from the object name by appending .yaml to it. The path where the file is
// removed from is set when the ReadLocalManifests function creates the
// ApiObjectStore.
func (aos *localFileApiObjectStore) RemoveResource(_ context.Context, obj runtime.Object) error {
	return os.Remove(getFileName(obj))
}

// ReadLocalManifests creates a new ApiObjectStore. It reads all files under path.
// The files are deserialized and validated.
func ReadLocalManifests(path string, options globalregistry.RegistryOptions) (*localFileApiObjectStore, error) {
	aos := &localFileApiObjectStore{
		path:    path,
		options: options,
	}
	aos.serializer = json.NewSerializerWithOptions(
		json.DefaultMetaFactory,
		scheme,
		scheme,
		json.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		})
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer dir.Close()
	entries, err := dir.ReadDir(16 * 1024)
	if err != nil {
		return nil, err
	}
	aos.store = make(map[schema.GroupVersionKind][]runtime.Object)
	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".yaml") {
			// skip the non-yaml files
			continue
		}
		f, err := os.Open(filepath.Join(path, entry.Name()))
		if err != nil {
			return nil, err
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}
		o, gvk, err := aos.serializer.Decode(b, nil, nil)
		if err != nil {
			// This is not a valid Kubernetes resource. Let's skip it.
			logger.V(-1).Info("file is not a valid resource",
				"error", err.Error(),
				"filename", entry.Name(),
			)
			continue
		}
		err = validateObjects(o, gvk)
		if err != nil {
			return nil, fmt.Errorf("validation error during inspecting %s:\n %w", entry.Name(), err)
		}
		objects, found := aos.store[*gvk]
		if found {
			aos.store[*gvk] = append(objects, o)
		} else {
			aos.store[*gvk] = []runtime.Object{o}
		}
	}
	err = aos.validate()
	if err != nil {
		return nil, err
	}
	return aos, nil
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

// validate performs all validations that require the full context, i.e. all
// resources parsed.
func (aos *localFileApiObjectStore) validate() error {
	registries := aos.GetRegistries(context.Background())
	projects := aos.GetProjects(context.Background())
	scanners := aos.GetScanners(context.Background())

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
	if project.Spec.Trigger.Schedule != "" &&
		project.Spec.Trigger.TriggerType() != api.CronReplicationTriggerType {
		return ErrValidationReplicationRuleWrongSchedule
	}
	return err
}

// validateObjects perform the CRD level validation on each object.
func validateObjects(o runtime.Object, gvk *schema.GroupVersionKind) error {
	var results *validate.Result

	switch gvk.Kind {
	case "Registry":
		results = api.RegistryValidator.Validate(o)
	case "Project":
		results = api.ProjectValidator.Validate(o)
	case "Scanner":
		results = api.ScannerValidator.Validate(o)
	default:
		// We have parsed a Kubernetes resource which is not our kind.
		// Don't validate it.
		return nil
	}

	if results.HasErrors() {
		return results.AsError()
	}
	switch gvk.Kind {
	case "Project":
		return checkProject(o.(*api.Project))
	default:
		return nil
	}
}

// GetRegistries returns the parsed registries as API objects.
func (aos *localFileApiObjectStore) GetRegistries(ctx context.Context) []*api.Registry {
	registryObjects, found := aos.store[api.SchemeGroupVersion.WithKind("Registry")]
	if !found {
		return []*api.Registry{}
	}
	registries := make([]*api.Registry, len(registryObjects))
	for i, reg := range registryObjects {
		registries[i] = reg.(*api.Registry)
	}
	return registries
}

// GetProjects returns the parsed projects as API objects.
func (aos *localFileApiObjectStore) GetProjects(context.Context) []*api.Project {
	projectObjects, found := aos.store[api.SchemeGroupVersion.WithKind("Project")]
	if !found {
		return []*api.Project{}
	}
	projects := make([]*api.Project, len(projectObjects))
	for i, reg := range projectObjects {
		projects[i] = reg.(*api.Project)
	}
	return projects
}

// GetScanners returns the parsed scanners as API objects.
func (aos *localFileApiObjectStore) GetScanners(context.Context) []*api.Scanner {
	scannerObjects, found := aos.store[api.SchemeGroupVersion.WithKind("Scanner")]
	if !found {
		return []*api.Scanner{}
	}
	scanners := make([]*api.Scanner, len(scannerObjects))
	for i, reg := range scannerObjects {
		scanners[i] = reg.(*api.Scanner)
	}
	return scanners
}

// GetGlobalRegistryOptions returns the ApiObjectStore related CLI options of an
// apply.
func (aos *localFileApiObjectStore) GetGlobalRegistryOptions() globalregistry.RegistryOptions {
	return aos.options
}

func (aos *localFileApiObjectStore) GetLogger() logr.Logger {
	return logger
}

func (aos *localFileApiObjectStore) UpdateRegistryStatus(ctx context.Context, reg *api.Registry) error {
	// We don't persist the status for filesystem based resources.
	return nil
}

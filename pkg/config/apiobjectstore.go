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
	"fmt"
	"io"
	"os"
	"path/filepath"

	api "github.com/kubermatic-labs/registryman/pkg/apis/globalregistry/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/config/registry"
	_ "github.com/kubermatic-labs/registryman/pkg/harbor"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/kube-openapi/pkg/validation/validate"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
	}
	err := api.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	scheme.AddKnownTypeWithName(secret.GroupVersionKind(), secret)
}

type ApiObjectStore struct {
	store      map[schema.GroupVersionKind][]runtime.Object
	serializer *json.Serializer
}

func (aos *ApiObjectStore) GetSerializer() *json.Serializer {
	return aos.serializer
}

func ReadManifests(path string) (*ApiObjectStore, error) {
	aos := &ApiObjectStore{}
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
			return nil, err
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
	err = aos.Validate()
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return aos, nil
}

func checkGlobalRegistryCount(registries []*api.Registry) error {
	globalRegistries := make([]string, 0)
	for _, registry := range registries {
		if registry.Spec.Role == "GlobalHub" {
			globalRegistries = append(globalRegistries, registry.Name)
		}
	}
	if len(globalRegistries) >= 2 {
		for _, registry := range globalRegistries {
			logger.V(-2).Info("Multiple Global Registries found",
				"registry_name", registry)
		}
		return fmt.Errorf("%w", ValidationErrorMultipleGlobalRegistries)
	}
	return nil
}

func checkLocalRegistryNamesInProjects(registries []*api.Registry, projects []*api.Project) error {
	localRegistries := make([]string, 0)
	for _, registry := range registries {
		if registry.Spec.Role == "Local" {
			localRegistries = append(localRegistries, registry.Name)
		}
	}
	invalidProjects := make([]string, 0)
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
					invalidProjects = append(invalidProjects, project.Name)
					logger.V(-2).Info("Local registry not exists",
						"project_name", project.Name,
						"registry_name", localRegistry)
				}
				localRegistryExists = false
			}
		}
	}

	if len(invalidProjects) > 0 {
		return fmt.Errorf("%w", ValidationErrorInvalidLocalRegistryInProject)
	}
	return nil
}

func (aos *ApiObjectStore) Validate() error {
	provider := aos.ApiProvider()
	registries := provider.GetRegistries()
	projects := provider.GetProjects()

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

	return nil
}

func validateObjects(o runtime.Object, gvk *schema.GroupVersionKind) error {
	var results *validate.Result

	switch gvk.Kind {
	case "Registry":
		results = api.RegistryValidator.Validate(o)
	case "Project":
		results = api.ProjectValidator.Validate(o)
	default:
		return fmt.Errorf("%s Kind is not supported", gvk.Kind)
	}

	if results.HasErrors() {
		return results.AsError()
	}
	return nil
}

type ApiProvider ApiObjectStore

func (aos *ApiObjectStore) ApiProvider() *ApiProvider {
	return (*ApiProvider)(aos)
}

func (apip *ApiProvider) GetRegistries() []*api.Registry {
	registryObjects, found := apip.store[api.SchemeGroupVersion.WithKind("Registry")]
	if !found {
		return []*api.Registry{}
	}
	registries := make([]*api.Registry, len(registryObjects))
	for i, reg := range registryObjects {
		registries[i] = reg.(*api.Registry)
	}
	return registries
}

func (apip *ApiProvider) GetProjects() []*api.Project {
	projectObjects, found := apip.store[api.SchemeGroupVersion.WithKind("Project")]
	if !found {
		return []*api.Project{}
	}
	projects := make([]*api.Project, len(projectObjects))
	for i, reg := range projectObjects {
		projects[i] = reg.(*api.Project)
	}
	return projects
}

type ExpectedProvider ApiObjectStore

func (aos *ApiObjectStore) ExpectedProvider() *ExpectedProvider {
	return (*ExpectedProvider)(aos)
}

func (expp *ExpectedProvider) GetRegistries() []*registry.Registry {
	registryObjects, found := expp.store[api.SchemeGroupVersion.WithKind("Registry")]
	if !found {
		panic("cannot get registry objects")
	}
	registries := make([]*registry.Registry, len(registryObjects))
	for i, reg := range registryObjects {
		apiRegistry, ok := reg.(*api.Registry)
		if !ok {
			panic("cannot assert registry object")
		}
		registries[i] = registry.New(apiRegistry, (*ApiObjectStore)(expp).ApiProvider())
	}
	return registries
}

func (expp *ExpectedProvider) GetRegistryByName(name string) *registry.Registry {
	registryObjects, found := expp.store[api.SchemeGroupVersion.WithKind("Registry")]
	if !found {
		panic("cannot get registry objects")
	}
	for _, reg := range registryObjects {
		apiRegistry, ok := reg.(*api.Registry)
		if !ok {
			panic("cannot assert registry object")
		}
		if apiRegistry.GetName() == name {
			return registry.New(apiRegistry, (*ApiObjectStore)(expp).ApiProvider())
		}
	}
	return nil
}

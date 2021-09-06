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
	return aos, nil
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

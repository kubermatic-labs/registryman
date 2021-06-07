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
		objects, found := aos.store[*gvk]
		if found {
			aos.store[*gvk] = append(objects, o)
		} else {
			aos.store[*gvk] = []runtime.Object{o}
		}
	}
	return aos, nil
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

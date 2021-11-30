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

package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/config"
	admissionV1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8sjson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sync"
)

var (
	serializer     *k8sjson.Serializer
	aos            config.ApiObjectStore
	aosOnce        sync.Once
	serializerOnce sync.Once
)

func getSerializer() *k8sjson.Serializer {
	serializerOnce.Do(func() {
		scheme := runtime.NewScheme()
		err := api.AddToScheme(scheme)
		if err != nil {
			panic(err)
		}
		serializer = k8sjson.NewSerializerWithOptions(
			k8sjson.DefaultMetaFactory,
			scheme,
			scheme,
			k8sjson.SerializerOptions{
				Yaml:   false,
				Pretty: false,
				Strict: true,
			})
	})
	return serializer
}

func getAos() config.ApiObjectStore {
	aosOnce.Do(func() {
		var err error
		aos, _, err = config.ConnectToKube(nil)
		if err != nil {
			panic(err)
		}
	})
	return aos
}

type mockApiObjestStore struct {
	config.ApiObjectStore
	addedRegistry   *api.Registry
	addedProject    *api.Project
	addedScanner    *api.Scanner
	removedRegistry *api.Registry
	removedProject  *api.Project
	removedScanner  *api.Scanner
}

func (maos *mockApiObjestStore) GetRegistries(ctx context.Context) []*api.Registry {
	var result []*api.Registry
	if maos.addedRegistry != nil {
		result = []*api.Registry{maos.addedRegistry}
	} else {
		result = []*api.Registry{}
	}
	var removedRegistryName string
	var removedRegistryNamespace string
	if maos.removedRegistry != nil {
		removedRegistryName = maos.removedRegistry.GetName()
		removedRegistryNamespace = maos.removedRegistry.GetNamespace()
	}
	projects := maos.ApiObjectStore.GetRegistries(ctx)
	for i, registry := range maos.ApiObjectStore.GetRegistries(ctx) {
		if registry.GetName() != removedRegistryName ||
			registry.GetNamespace() != removedRegistryNamespace {
			result = append(result, projects[i])
		}
	}
	return result
}

func (maos *mockApiObjestStore) GetProjects(ctx context.Context) []*api.Project {
	var result []*api.Project
	if maos.addedProject != nil {
		result = []*api.Project{maos.addedProject}
	} else {
		result = []*api.Project{}
	}
	var removedProjectName string
	var removedProjectNamespace string
	if maos.removedProject != nil {
		removedProjectName = maos.removedProject.GetName()
		removedProjectNamespace = maos.removedProject.GetNamespace()
	}
	projects := maos.ApiObjectStore.GetProjects(ctx)
	for i, project := range maos.ApiObjectStore.GetProjects(ctx) {
		if project.GetName() != removedProjectName ||
			project.GetNamespace() != removedProjectNamespace {
			result = append(result, projects[i])
		}
	}
	return result
}

func (maos *mockApiObjestStore) GetScanners(ctx context.Context) []*api.Scanner {
	var result []*api.Scanner
	if maos.addedScanner != nil {
		result = []*api.Scanner{maos.addedScanner}
	} else {
		result = []*api.Scanner{}
	}
	var removedScannerName string
	var removedScannerNamespace string
	if maos.removedScanner != nil {
		removedScannerName = maos.removedScanner.GetName()
		removedScannerNamespace = maos.removedScanner.GetNamespace()
	}
	projects := maos.ApiObjectStore.GetScanners(ctx)
	for i, scanner := range maos.ApiObjectStore.GetScanners(ctx) {
		if scanner.GetName() != removedScannerName ||
			scanner.GetNamespace() != removedScannerNamespace {
			result = append(result, projects[i])
		}
	}
	return result
}

func mockAOSWithRegistry(reg *api.Registry) *mockApiObjestStore {
	return &mockApiObjestStore{
		ApiObjectStore: getAos(),
		addedRegistry:  reg,
	}
}

func mockAOSWithoutRegistry(reg *api.Registry) *mockApiObjestStore {
	return &mockApiObjestStore{
		ApiObjectStore:  getAos(),
		removedRegistry: reg,
	}
}

func mockAOSWithUpdatedRegistry(oldRegistry, newRegistry *api.Registry) *mockApiObjestStore {
	return &mockApiObjestStore{
		ApiObjectStore:  getAos(),
		addedRegistry:   newRegistry,
		removedRegistry: oldRegistry,
	}
}

func mockAOSWithProject(proj *api.Project) *mockApiObjestStore {
	return &mockApiObjestStore{
		ApiObjectStore: getAos(),
		addedProject:   proj,
	}
}

func mockAOSWithoutProject(proj *api.Project) *mockApiObjestStore {
	return &mockApiObjestStore{
		ApiObjectStore: getAos(),
		removedProject: proj,
	}
}

func mockAOSWithUpdatedProject(oldProject, newProject *api.Project) *mockApiObjestStore {
	return &mockApiObjestStore{
		ApiObjectStore: getAos(),
		addedProject:   newProject,
		removedProject: oldProject,
	}
}

func mockAOSWithScanner(scanner *api.Scanner) *mockApiObjestStore {
	return &mockApiObjestStore{
		ApiObjectStore: getAos(),
		addedScanner:   scanner,
	}
}

func mockAOSWithoutScanner(scanner *api.Scanner) *mockApiObjestStore {
	return &mockApiObjestStore{
		ApiObjectStore: getAos(),
		removedScanner: scanner,
	}
}

func mockAOSWithUpdatedScanner(oldScanner, newScanner *api.Scanner) *mockApiObjestStore {
	return &mockApiObjestStore{
		ApiObjectStore: getAos(),
		addedScanner:   newScanner,
		removedScanner: oldScanner,
	}
}

func AdmissionRequestHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("admission request handler invoked",
		"method", r.Method,
		"URL", r.URL.String(),
		"Headers", r.Header,
	)
	buf := new(bytes.Buffer)

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		logger.Error(err, "cannot read HTTP response body")
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(buf.Bytes()))

	var admissionRev admissionV1.AdmissionReview
	err = decoder.Decode(&admissionRev)
	if err != nil {
		logger.Error(err, "invalid request body",
			"body", buf.String(),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch admissionRev.Request.Operation {
	default:
		logger.V(-2).Info("unknown operation",
			"operation", admissionRev.Request.Operation,
		)
	case admissionV1.Create:
		createAdmissionHandler(w, &admissionRev)
	case admissionV1.Delete:
		deleteAdmissionHandler(w, &admissionRev)
	case admissionV1.Update:
		updateAdmissionHandler(w, &admissionRev)
	}

}

func validateConsistency(w http.ResponseWriter, aos config.ApiObjectStore, admissionRev *admissionV1.AdmissionReview) {
	encoder := json.NewEncoder(w)

	err := config.ValidateConsistency(aos)
	if err != nil {
		logger.Info("rejecting validation request",
			"reason", err.Error(),
		)
		admissionRev.Response = &admissionV1.AdmissionResponse{
			UID:     admissionRev.Request.UID,
			Allowed: false,
			Result: &metav1.Status{
				Message: err.Error(),
				Code:    403,
			},
		}
	} else {
		logger.Info("accepting validation request")
		admissionRev.Response = &admissionV1.AdmissionResponse{
			UID:     admissionRev.Request.UID,
			Allowed: true,
		}
	}

	err = encoder.Encode(&admissionRev)
	if err != nil {
		logger.Error(err, "error encoding response body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func createAdmissionHandler(w http.ResponseWriter, admissionRev *admissionV1.AdmissionReview) {
	o, gvk, err := getSerializer().Decode(admissionRev.Request.Object.Raw, nil, nil)
	if err != nil {
		logger.V(-2).Info("unknwon resource",
			"error", err.Error(),
			"raw", string(admissionRev.Request.Object.Raw),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logger.Info("adding object",
		"gvk", gvk,
	)
	var aos config.ApiObjectStore
	switch admissionRev.Request.Kind {
	case metav1.GroupVersionKind{
		Group:   api.GroupName,
		Version: api.GroupVersion.Version,
		Kind:    "Registry",
	}:
		reg, ok := o.(*api.Registry)
		if !ok {
			logger.V(-2).Info("registry type mismatch")
			http.Error(w, "Registry type mismatch", http.StatusBadRequest)
			return
		}
		aos = mockAOSWithRegistry(reg)
	case metav1.GroupVersionKind{
		Group:   api.GroupName,
		Version: api.GroupVersion.Version,
		Kind:    "Project",
	}:
		proj, ok := o.(*api.Project)
		if !ok {
			logger.V(-2).Info("project type mismatch",
				"request", admissionRev.Request,
			)
			http.Error(w, "Project type mismatch", http.StatusBadRequest)
			return
		}
		aos = mockAOSWithProject(proj)
	case metav1.GroupVersionKind{
		Group:   api.GroupName,
		Version: api.GroupVersion.Version,
		Kind:    "Scanner",
	}:
		scanner, ok := o.(*api.Scanner)
		if !ok {
			logger.V(-2).Info("scanner type mismatch")
			http.Error(w, "Scanner type mismatch", http.StatusBadRequest)
			return
		}
		aos = mockAOSWithScanner(scanner)
	}
	validateConsistency(w, aos, admissionRev)
}

func deleteAdmissionHandler(w http.ResponseWriter, admissionRev *admissionV1.AdmissionReview) {
	o, gvk, err := getSerializer().Decode(admissionRev.Request.OldObject.Raw, nil, nil)
	if err != nil {
		logger.V(-2).Info("unknwon resource",
			"error", err.Error(),
			"raw", string(admissionRev.Request.Object.Raw),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logger.Info("removing object",
		"gvk", gvk,
	)
	var aos config.ApiObjectStore
	switch admissionRev.Request.Kind {
	case metav1.GroupVersionKind{
		Group:   api.GroupName,
		Version: api.GroupVersion.Version,
		Kind:    "Registry",
	}:
		reg, ok := o.(*api.Registry)
		if !ok {
			logger.V(-2).Info("registry type mismatch")
			http.Error(w, "Registry type mismatch", http.StatusBadRequest)
			return
		}
		aos = mockAOSWithoutRegistry(reg)
	case metav1.GroupVersionKind{
		Group:   api.GroupName,
		Version: api.GroupVersion.Version,
		Kind:    "Project",
	}:
		proj, ok := o.(*api.Project)
		if !ok {
			logger.V(-2).Info("project type mismatch",
				"request", admissionRev.Request,
			)
			http.Error(w, "Project type mismatch", http.StatusBadRequest)
			return
		}
		aos = mockAOSWithoutProject(proj)
	case metav1.GroupVersionKind{
		Group:   api.GroupName,
		Version: api.GroupVersion.Version,
		Kind:    "Scanner",
	}:
		scanner, ok := o.(*api.Scanner)
		if !ok {
			logger.V(-2).Info("scanner type mismatch")
			http.Error(w, "Scanner type mismatch", http.StatusBadRequest)
			return
		}
		aos = mockAOSWithoutScanner(scanner)
	}
	validateConsistency(w, aos, admissionRev)
}

func updateAdmissionHandler(w http.ResponseWriter, admissionRev *admissionV1.AdmissionReview) {
	o, gvk, err := getSerializer().Decode(admissionRev.Request.Object.Raw, nil, nil)
	if err != nil {
		logger.V(-2).Info("unknwon resource",
			"error", err.Error(),
			"raw", string(admissionRev.Request.Object.Raw),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	oldO, _, err := getSerializer().Decode(admissionRev.Request.OldObject.Raw, nil, nil)
	if err != nil {
		logger.V(-2).Info("unknwon resource",
			"error", err.Error(),
			"raw", string(admissionRev.Request.Object.Raw),
		)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logger.Info("updating object",
		"gvk", gvk,
	)
	var aos config.ApiObjectStore
	switch admissionRev.Request.Kind {
	case metav1.GroupVersionKind{
		Group:   api.GroupName,
		Version: api.GroupVersion.Version,
		Kind:    "Registry",
	}:
		reg, ok := o.(*api.Registry)
		if !ok {
			logger.V(-2).Info("registry type mismatch")
			http.Error(w, "Registry type mismatch", http.StatusBadRequest)
			return
		}
		oldreg, ok := oldO.(*api.Registry)
		if !ok {
			logger.V(-2).Info("registry type mismatch")
			http.Error(w, "Registry type mismatch", http.StatusBadRequest)
			return
		}
		aos = mockAOSWithUpdatedRegistry(oldreg, reg)
	case metav1.GroupVersionKind{
		Group:   api.GroupName,
		Version: api.GroupVersion.Version,
		Kind:    "Project",
	}:
		proj, ok := o.(*api.Project)
		if !ok {
			logger.V(-2).Info("project type mismatch")
			http.Error(w, "Project type mismatch", http.StatusBadRequest)
			return
		}
		oldproj, ok := oldO.(*api.Project)
		if !ok {
			logger.V(-2).Info("project type mismatch")
			http.Error(w, "Project type mismatch", http.StatusBadRequest)
			return
		}
		aos = mockAOSWithUpdatedProject(oldproj, proj)
	case metav1.GroupVersionKind{
		Group:   api.GroupName,
		Version: api.GroupVersion.Version,
		Kind:    "Scanner",
	}:
		scanner, ok := o.(*api.Scanner)
		if !ok {
			logger.V(-2).Info("scanner type mismatch")
			http.Error(w, "Scanner type mismatch", http.StatusBadRequest)
			return
		}
		oldscanner, ok := oldO.(*api.Scanner)
		if !ok {
			logger.V(-2).Info("scanner type mismatch")
			http.Error(w, "Scanner type mismatch", http.StatusBadRequest)
			return
		}
		aos = mockAOSWithUpdatedScanner(oldscanner, scanner)
	}
	validateConsistency(w, aos, admissionRev)
}

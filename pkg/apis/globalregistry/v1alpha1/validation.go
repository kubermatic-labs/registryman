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
package v1alpha1

import (
	_ "embed"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/kube-openapi/pkg/validation/validate"

	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

//go:embed registryman.kubermatic.com_registries.yaml
var registryCRDYaml []byte

//go:embed registryman.kubermatic.com_projects.yaml
var projectCRDYaml []byte

//go:embed registryman.kubermatic.com_scanners.yaml
var scannerCRDYaml []byte

// RegistryCRD is the CustomResourceDefinition representation of the generated
// Registry CRD yaml.
var RegistryCRD *apiext.CustomResourceDefinition

// ProjectCRD is the CustomResourceDefinition representation of the generated
// Project CRD yaml.
var ProjectCRD *apiext.CustomResourceDefinition

// ScannerCRD is the CustomResourceDefinition representation of the generated
// Scanner CRD yaml.
var ScannerCRD *apiext.CustomResourceDefinition

var RegistryValidator *validate.SchemaValidator
var ProjectValidator *validate.SchemaValidator
var ScannerValidator *validate.SchemaValidator

func init() {
	scheme := runtime.NewScheme()
	err := apiextv1.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	err = apiext.AddToScheme(scheme)
	if err != nil {
		panic(err)
	}
	serializer := json.NewSerializerWithOptions(json.DefaultMetaFactory, scheme, scheme, json.SerializerOptions{
		Yaml:   true,
		Pretty: true,
		Strict: true,
	})

	registryCRDv1, _, err := serializer.Decode(registryCRDYaml, nil, nil)
	if err != nil {
		panic(err)
	}

	registryCRD, err := scheme.ConvertToVersion(registryCRDv1, apiext.SchemeGroupVersion)
	if err != nil {
		panic(err)
	}

	var ok bool
	RegistryCRD, ok = registryCRD.(*apiext.CustomResourceDefinition)
	if !ok {
		panic("registry CRD yaml is not a valid CustomResourceDefinition")
	}

	projectCRDv1, _, err := serializer.Decode(projectCRDYaml, nil, nil)
	if err != nil {
		panic(err)
	}

	projectCRD, err := scheme.ConvertToVersion(projectCRDv1, apiext.SchemeGroupVersion)
	if err != nil {
		panic(err)
	}

	ProjectCRD, ok = projectCRD.(*apiext.CustomResourceDefinition)
	if !ok {
		panic("project CRD yaml is not a valid CustomResourceDefinition")
	}

	scannerCRDv1, _, err := serializer.Decode(scannerCRDYaml, nil, nil)
	if err != nil {
		panic(err)
	}

	scannerCRD, err := scheme.ConvertToVersion(scannerCRDv1, apiext.SchemeGroupVersion)
	if err != nil {
		panic(err)
	}

	ScannerCRD, ok = scannerCRD.(*apiext.CustomResourceDefinition)
	if !ok {
		panic("scanner CRD yaml is not a valid CustomResourceDefinition")
	}

	RegistryValidator, _, err = validation.NewSchemaValidator(RegistryCRD.Spec.Validation)
	if err != nil {
		panic(err)
	}

	ProjectValidator, _, err = validation.NewSchemaValidator(ProjectCRD.Spec.Validation)
	if err != nil {
		panic(err)
	}

	ScannerValidator, _, err = validation.NewSchemaValidator(ScannerCRD.Spec.Validation)
	if err != nil {
		panic(err)
	}
}

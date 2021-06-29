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

package v1alpha1_test

import (
	"fmt"
	"io"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/kubermatic-labs/registryman/pkg/apis/registryman.kubermatic.com/v1alpha1"
	"k8s.io/kube-openapi/pkg/validation/errors"
)

var serializer *json.Serializer

func init() {
	scheme := runtime.NewScheme()
	err := v1alpha1.Install(scheme)
	if err != nil {
		panic(err)
	}

	serializer = json.NewSerializerWithOptions(
		json.DefaultMetaFactory,
		scheme, scheme,
		json.SerializerOptions{
			Yaml:   true,
			Pretty: true,
			Strict: true,
		})
}

func objectFromFile(filename string) (runtime.Object, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	bs, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	registry, _, err := serializer.Decode(bs, nil, nil)
	return registry, err
}

var _ = Describe("Validation", func() {
	It("can validate valid Registry resources", func() {
		registry, err := objectFromFile("testdata/global-registry.yaml")
		Expect(err).ToNot(HaveOccurred())

		results := v1alpha1.RegistryValidator.Validate(registry)
		if results.HasErrors() {
			fmt.Fprintln(GinkgoWriter, results.AsError().Error())
		}
		Expect(results.HasErrorsOrWarnings()).To(BeFalse())

		registry, err = objectFromFile("testdata/local-registry.yaml")
		Expect(err).ToNot(HaveOccurred())

		results = v1alpha1.RegistryValidator.Validate(registry)
		if results.HasErrors() {
			fmt.Fprintln(GinkgoWriter, results.AsError().Error())
		}
		Expect(results.HasErrorsOrWarnings()).To(BeFalse())
	})

	It("can validate valid Project resources", func() {
		project, err := objectFromFile("testdata/local-project.yaml")
		Expect(err).ToNot(HaveOccurred())

		results := v1alpha1.ProjectValidator.Validate(project)
		if results.HasErrors() {
			fmt.Fprintln(GinkgoWriter, results.AsError().Error())
		}
		Expect(results.HasErrorsOrWarnings()).To(BeFalse())
	})

	It("will fail for invalid Registry resources", func() {
		registry, err := objectFromFile("testdata/registry-wrong-apiendpoint.yaml")
		Expect(err).ToNot(HaveOccurred())

		results := v1alpha1.RegistryValidator.Validate(registry)
		Expect(results.HasErrorsOrWarnings()).To(BeTrue())
		if results.HasErrors() {
			fmt.Fprintln(GinkgoWriter, results.AsError().Error())
			c, ok := results.AsError().(*errors.CompositeError)
			Expect(ok).To(BeTrue())
			Expect(len(c.Errors)).Should(Equal(1))
			for _, e := range c.Errors {
				v := e.(*errors.Validation)
				Expect(v.Name).To(Equal("spec.apiEndpoint"))
				Expect(v.Code()).Should(Equal(int32(errors.PatternFailCode)))
			}
		}
	})
	It("will fail for invalid Scanner resources", func() {
		scanner, err := objectFromFile("testdata/scanner-wrong-url.yaml")
		Expect(err).ToNot(HaveOccurred())

		results := v1alpha1.ScannerValidator.Validate(scanner)
		Expect(results.HasErrorsOrWarnings()).To(BeTrue())
		if results.HasErrors() {
			fmt.Fprintln(GinkgoWriter, results.AsError().Error())
			c, ok := results.AsError().(*errors.CompositeError)
			Expect(ok).To(BeTrue())
			Expect(len(c.Errors)).Should(Equal(1))
			for _, e := range c.Errors {
				v := e.(*errors.Validation)
				Expect(v.Name).To(Equal("spec.url"))
				Expect(v.Code()).Should(Equal(int32(errors.PatternFailCode)))
			}
		}
	})
})

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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

const (
	apiVersion   = "registryman.kubermatic.com/v1alpha1"
	registryKind = "Registry"
	testdataDir  = "testdata"
)

var _ = Describe("Registry", func() {
	var registry *api.Registry
	var scheme *runtime.Scheme
	var serializer runtime.Serializer
	BeforeEach(func() {
		registry = &api.Registry{}
		gvk := schema.FromAPIVersionAndKind(apiVersion, registryKind)
		registry.SetGroupVersionKind(gvk)
		scheme = runtime.NewScheme()
		Expect(scheme).NotTo(BeNil())
		scheme.AddKnownTypeWithName(gvk, registry)
		serializer = json.NewSerializerWithOptions(
			json.DefaultMetaFactory,
			scheme,
			scheme,
			json.SerializerOptions{
				Yaml:   true,
				Pretty: true,
				Strict: true,
			})
		Expect(serializer).ToNot(BeNil())
	})
	Context("when reading global-registry.yaml", func() {
		It("can be decoded", func() {
			b, err := openTestFile("global-registry.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(b).ToNot(BeNil())
			o, _, err := serializer.Decode(b, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			r, ok := o.(*api.Registry)
			Expect(ok).To(BeTrue())
			Expect(r.GetName()).To(Equal("global"))
			Expect(r.Spec.Provider).To(Equal("harbor"))
			Expect(r.Spec.Role).To(Equal("GlobalHub"))
			Expect(r.Spec.APIEndpoint).To(Equal("http://core.harbor-1.demo"))
			Expect(r.Spec.Username).To(Equal("admin"))
			Expect(r.Spec.Password).To(Equal("admin"))
		})
	})
})

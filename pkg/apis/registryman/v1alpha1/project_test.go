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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
)

const (
	projectKind = "Project"
)

var _ = Describe("Project", func() {
	var project *api.Project
	var scheme *runtime.Scheme
	var serializer runtime.Serializer
	BeforeEach(func() {
		project = &api.Project{}
		gvk := schema.FromAPIVersionAndKind(apiVersion, projectKind)
		project.SetGroupVersionKind(gvk)
		scheme = runtime.NewScheme()
		Expect(scheme).NotTo(BeNil())
		scheme.AddKnownTypeWithName(gvk, project)
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
	Context("when reading global-project.yaml", func() {
		It("can be decoded", func() {
			b, err := openTestFile("global-project.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(b).ToNot(BeNil())
			o, _, err := serializer.Decode(b, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			r, ok := o.(*api.Project)
			Expect(ok).To(BeTrue())
			Expect(r.GetName()).To(Equal("ubuntu"))
			Expect(r.Spec.Type).To(Equal(api.GlobalProjectType))
			Expect(len(r.Spec.LocalRegistries)).To(Equal(0))
			Expect(len(r.Spec.Members)).To(Equal(4))
			Expect(r.Spec.Members[0].Name).To(Equal("alpha"))
			Expect(r.Spec.Members[0].Type).To(Equal(api.UserMemberType))
			Expect(r.Spec.Members[0].Role).To(Equal(api.MaintainerRole))
			Expect(r.Spec.Members[1].Name).To(Equal("beta"))
			Expect(r.Spec.Members[1].Type).To(Equal(api.UserMemberType))
			Expect(r.Spec.Members[1].Role).To(Equal(api.DeveloperRole))
			Expect(r.Spec.Members[2].Name).To(Equal("ci-robot"))
			Expect(r.Spec.Members[2].Type).To(Equal(api.RobotMemberType))
			Expect(r.Spec.Members[2].Role).To(Equal(api.PushOnlyRole))
			Expect(r.Spec.Members[3].Name).To(Equal("project-admins"))
			Expect(r.Spec.Members[3].Type).To(Equal(api.GroupMemberType))
			Expect(r.Spec.Members[3].Role).To(Equal(api.ProjectAdminRole))
		})
	})
	Context("when reading local-project.yaml", func() {
		It("can be decoded", func() {
			b, err := openTestFile("local-project.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(b).ToNot(BeNil())
			o, _, err := serializer.Decode(b, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			r, ok := o.(*api.Project)
			Expect(ok).To(BeTrue())
			Expect(r.GetName()).To(Equal("node"))
			Expect(r.Spec.Type).To(Equal(api.LocalProjectType))
			Expect(len(r.Spec.LocalRegistries)).To(Equal(1))
			Expect(r.Spec.LocalRegistries[0]).To(Equal("local"))
			Expect(len(r.Spec.Members)).To(Equal(4))
			Expect(r.Spec.Members[0].Name).To(Equal("alpha"))
			Expect(r.Spec.Members[0].Type).To(Equal(api.UserMemberType))
			Expect(r.Spec.Members[0].Role).To(Equal(api.MaintainerRole))
			Expect(r.Spec.Members[1].Name).To(Equal("beta"))
			Expect(r.Spec.Members[1].Type).To(Equal(api.UserMemberType))
			Expect(r.Spec.Members[1].Role).To(Equal(api.DeveloperRole))
			Expect(r.Spec.Members[2].Name).To(Equal("ci-robot"))
			Expect(r.Spec.Members[2].Type).To(Equal(api.RobotMemberType))
			Expect(r.Spec.Members[2].Role).To(Equal(api.PushOnlyRole))
			Expect(r.Spec.Members[3].Name).To(Equal("project-admins"))
			Expect(r.Spec.Members[3].Type).To(Equal(api.GroupMemberType))
			Expect(r.Spec.Members[3].Role).To(Equal(api.ProjectAdminRole))
		})
	})
})

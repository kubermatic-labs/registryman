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
	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Options", func() {
	Context("when getting registries with forceDelete options", func() {
		It("gets the correct value when ForceDelete is false", func() {
			b, err := openTestFile("test_options/global-registry-forcedelete.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(b).ToNot(BeNil())
			o, _, err := serializer.Decode(b, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			r, ok := o.(*api.Registry)
			Expect(ok).To(BeTrue())
			Expect(r.Annotations).To(HaveKeyWithValue("registryman.kubermatic.com/forceDelete", "false"))
		})
		It("gets the correct value when ForceDelete is true", func() {
			b, err := openTestFile("test_options/local-registry-forcedelete.yaml")
			Expect(err).ToNot(HaveOccurred())
			Expect(b).ToNot(BeNil())
			o, _, err := serializer.Decode(b, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			r, ok := o.(*api.Registry)
			Expect(ok).To(BeTrue())
			Expect(r.Annotations).To(HaveKeyWithValue("registryman.kubermatic.com/forceDelete", "true"))
		})
	})
})

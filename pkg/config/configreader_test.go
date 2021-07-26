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

package config_test

import (
	"github.com/kubermatic-labs/registryman/pkg/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Configreader", func() {
	It("fails for invalid path", func() {
		m, err := config.ReadLocalManifests("nonexisting", nil)
		Expect(err).To(HaveOccurred())
		Expect(m).To(BeNil())
	})
	It("gets the correct values for api/testdata", func() {
		m, err := config.ReadLocalManifests("testdata", nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(m).ToNot(BeNil())
		registries := config.NewExpectedProvider(m).GetRegistries()
		Expect(len(registries)).To(Equal(2))
	})
	It("reads the registries and projects even with other yamls present", func() {
		m, err := config.ReadLocalManifests("testdata/test_other_yamls", nil)
		Expect(err).ToNot(HaveOccurred())
		Expect(m).ToNot(BeNil())
	})
})

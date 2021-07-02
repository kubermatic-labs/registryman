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

var _ = Describe("ProjectRepoName", func() {

	var manifests *config.ApiObjectStore
	BeforeEach(func() {
		var err error
		manifests, err = config.ReadManifests("testdata/test_projectreponame")
		Expect(err).ToNot(HaveOccurred())
		Expect(manifests).NotTo(BeNil())
	})

	It("throws an error for a non-existing project name", func() {
		projectRepoName, err := manifests.ApiProvider().ProjectRepoName("non-existing")
		Expect(err).Should(MatchError("project non-existing not found"))
		Expect(projectRepoName).To(Equal(""))
	})

	It("can generate the correct project repo name for a global project with a global registry", func() {
		projectRepoName, err := manifests.ApiProvider().ProjectRepoName("global")
		Expect(err).ToNot(HaveOccurred())
		Expect(projectRepoName).To(Equal("http://core.global.demo/global"))
	})

	It("can generate the correct project repo name for a local project with a global registry", func() {
		projectRepoName, err := manifests.ApiProvider().ProjectRepoName("local")
		Expect(err).ToNot(HaveOccurred())
		Expect(projectRepoName).To(Equal("http://core.local.demo/local"))
	})

})

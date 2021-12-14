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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kubermatic-labs/registryman/pkg/config"
)

var _ = Describe("Validation", func() {
	Context("when getting multiple global registries", func() {
		It("should error", func() {
			testDir := fmt.Sprintf("%s/test_multiple_global_registries", testdataDir)
			manifests, err := config.ReadLocalManifests(testDir, nil)
			Expect(manifests).NotTo(BeNil())
			Expect(err).To(Succeed())
			err = config.ValidateConsistency(manifests)
			Expect(err).Should(MatchError(config.ErrValidationMultipleGlobalRegistries))
		})
	})
	Context("when an artifactory registry has conflicting annotations", func() {
		It("should error", func() {
			testDir := fmt.Sprintf("%s/test_artifactory_annotations/annotation_conflict", testdataDir)
			manifests, err := config.ReadLocalManifests(testDir, nil)
			Expect(manifests).NotTo(BeNil())
			Expect(err).To(Succeed())
			err = config.ValidateConsistency(manifests)
			Expect(err).Should(MatchError(config.ErrValidationArtifactoryAnnotations))
		})
	})
	Context("when an artifactory registry has no annotations", func() {
		It("should error", func() {
			testDir := fmt.Sprintf("%s/test_artifactory_annotations/no_annotations", testdataDir)
			manifests, err := config.ReadLocalManifests(testDir, nil)
			Expect(manifests).NotTo(BeNil())
			Expect(err).To(Succeed())
			err = config.ValidateConsistency(manifests)
			Expect(err).Should(MatchError(config.ErrValidationArtifactoryAnnotations))
		})
	})
	Context("when an artifactory registry has one correct annotation", func() {
		It("should not fail", func() {
			testDir := fmt.Sprintf("%s/test_artifactory_annotations", testdataDir)
			manifests, err := config.ReadLocalManifests(testDir, nil)
			Expect(manifests).NotTo(BeNil())
			Expect(err).To(Succeed())
			err = config.ValidateConsistency(manifests)
			Expect(err).Should(BeNil())
		})
	})
	Context("when a project has invalid local registries", func() {
		It("should error", func() {
			testDir := fmt.Sprintf("%s/test_invalid_local_projects", testdataDir)
			manifests, err := config.ReadLocalManifests(testDir, nil)
			Expect(manifests).NotTo(BeNil())
			Expect(err).To(Succeed())
			err = config.ValidateConsistency(manifests)
			Expect(err).Should(MatchError(config.ErrValidationInvalidLocalRegistryInProject))
		})
	})
	Context("when there are multiple scanners with the same name", func() {
		It("should error", func() {
			testDir := fmt.Sprintf("%s/test_scannername_unique", testdataDir)
			manifests, err := config.ReadLocalManifests(testDir, nil)
			Expect(manifests).NotTo(BeNil())
			Expect(err).To(Succeed())
			err = config.ValidateConsistency(manifests)
			Expect(err).Should(MatchError(config.ErrValidationScannerNameNotUnique))
		})
	})
	Context("when there are multiple projects with the same name", func() {
		It("should error", func() {
			testDir := fmt.Sprintf("%s/test_projectname_unique", testdataDir)
			manifests, err := config.ReadLocalManifests(testDir, nil)
			Expect(manifests).NotTo(BeNil())
			Expect(err).To(Succeed())
			err = config.ValidateConsistency(manifests)
			Expect(err).Should(MatchError(config.ErrValidationProjectNameNotUnique))
		})
	})
	Context("when there are multiple registries with the same name", func() {
		It("should error", func() {
			testDir := fmt.Sprintf("%s/test_registryname_unique", testdataDir)
			manifests, err := config.ReadLocalManifests(testDir, nil)
			Expect(manifests).NotTo(BeNil())
			Expect(err).To(Succeed())
			err = config.ValidateConsistency(manifests)
			Expect(err).Should(MatchError(config.ErrValidationRegistryNameNotUnique))
		})
	})
	Context("when a project refers to a non-existing scanner", func() {
		It("should error", func() {
			testDir := fmt.Sprintf("%s/test_scannername_valid", testdataDir)
			manifests, err := config.ReadLocalManifests(testDir, nil)
			Expect(manifests).NotTo(BeNil())
			Expect(err).To(Succeed())
			err = config.ValidateConsistency(manifests)
			Expect(err).Should(MatchError(config.ErrValidationScannerNameReference))
		})
	})
})

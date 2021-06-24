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

package reconciler_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry/reconciler"
)

var _ = Describe("ScannerStatus", func() {
	It("returns no action for the same ScannerStatus values", func() {
		act := reconciler.ScannerStatus{
			Name: "scanner",
			URL:  "http://scanner.com",
		}
		exp := reconciler.ScannerStatus{
			Name: "scanner",
			URL:  "http://scanner.com",
		}
		actions := reconciler.CompareScannerStatuses("proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))

		By("empty ScannerStatus")
		act = reconciler.ScannerStatus{}
		exp = reconciler.ScannerStatus{}

		actions = reconciler.CompareScannerStatuses("proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))
	})

	It("can detect missing scanner configuration", func() {
		act := reconciler.ScannerStatus{}
		exp := reconciler.ScannerStatus{
			Name: "scanner",
			URL:  "http://scanner.com",
		}
		actions := reconciler.CompareScannerStatuses("proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"assigning scanner scanner to project proj",
		}))
	})

	It("can remove scanner configuration", func() {
		act := reconciler.ScannerStatus{
			Name: "scanner",
			URL:  "http://scanner.com",
		}
		exp := reconciler.ScannerStatus{}

		actions := reconciler.CompareScannerStatuses("proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"unassigning scanner scanner from project proj",
		}))
	})

	It("can detect scanner misconfiguration", func() {
		act := reconciler.ScannerStatus{
			Name: "scanner",
			URL:  "http://scanner.com",
		}
		exp := reconciler.ScannerStatus{
			Name: "scanner",
			URL:  "http://scanner-new.com",
		}

		actions := reconciler.CompareScannerStatuses("proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(2))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"unassigning scanner scanner from project proj",
			"assigning scanner scanner to project proj",
		}))

		act = reconciler.ScannerStatus{
			Name: "scanner",
			URL:  "http://scanner.com",
		}
		exp = reconciler.ScannerStatus{
			Name: "scanner-new",
			URL:  "http://scanner.com",
		}

		actions = reconciler.CompareScannerStatuses("proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(2))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"unassigning scanner scanner from project proj",
			"assigning scanner scanner-new to project proj",
		}))

		act = reconciler.ScannerStatus{
			Name: "scanner",
			URL:  "http://scanner.com",
		}
		exp = reconciler.ScannerStatus{
			Name: "scanner-new",
			URL:  "http://scanner-new.com",
		}

		actions = reconciler.CompareScannerStatuses("proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(2))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"unassigning scanner scanner from project proj",
			"assigning scanner scanner-new to project proj",
		}))
	})
})

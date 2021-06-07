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

	"bytes"
	"encoding/json"

	api "github.com/kubermatic-labs/registryman/pkg/apis/globalregistry/v1alpha1"
)

var _ = Describe("Memberrole", func() {
	It("can be stringified", func() {
		Expect(api.LimitedGuestRole.String()).To(Equal("LimitedGuest"))
		Expect(api.GuestRole.String()).To(Equal("Guest"))
		Expect(api.DeveloperRole.String()).To(Equal("Developer"))
		Expect(api.MaintainerRole.String()).To(Equal("Maintainer"))
		Expect(api.ProjectAdminRole.String()).To(Equal("ProjectAdmin"))
		Expect(api.PushOnlyRole.String()).To(Equal("PushOnly"))
		Expect(api.PullOnlyRole.String()).To(Equal("PullOnly"))
		Expect(api.PullAndPushRole.String()).To(Equal("PullAndPush"))
		Expect(func() { _ = api.MemberRole(-1).String() }).To(Panic())
	})

	It("can be json encoded", func() {
		s := []api.MemberRole{
			api.LimitedGuestRole,
			api.GuestRole,
			api.DeveloperRole,
			api.MaintainerRole,
			api.ProjectAdminRole,
			api.PushOnlyRole,
			api.PullOnlyRole,
			api.PullAndPushRole,
		}
		b := &bytes.Buffer{}
		err := json.NewEncoder(b).Encode(s)
		Expect(err).ToNot(HaveOccurred())
		Expect(b.String()).To(Equal("[\"LimitedGuest\",\"Guest\",\"Developer\",\"Maintainer\",\"ProjectAdmin\",\"PushOnly\",\"PullOnly\",\"PullAndPush\"]\n"))
	})
	It("can be json decoded", func() {
		var s []api.MemberRole
		b := bytes.NewBufferString("[\"LimitedGuest\",\"Guest\",\"Developer\",\"Maintainer\",\"ProjectAdmin\",\"PushOnly\",\"PullOnly\",\"PullAndPush\"]\n")
		err := json.NewDecoder(b).Decode(&s)
		Expect(err).ToNot(HaveOccurred())
		Expect(s).To(Equal([]api.MemberRole{
			api.LimitedGuestRole,
			api.GuestRole,
			api.DeveloperRole,
			api.MaintainerRole,
			api.ProjectAdminRole,
			api.PushOnlyRole,
			api.PullOnlyRole,
			api.PullAndPushRole,
		}))
	})
})

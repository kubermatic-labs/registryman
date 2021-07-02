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
	"bytes"
	"encoding/json"

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman.kubermatic.com/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Membertype", func() {
	It("can be stringified", func() {
		Expect(api.UserMemberType.String()).To(Equal("User"))
		Expect(api.GroupMemberType.String()).To(Equal("Group"))
		Expect(api.RobotMemberType.String()).To(Equal("Robot"))
		Expect(func() { _ = api.MemberRole(-1).String() }).To(Panic())
	})
	It("can be json encoded", func() {
		s := []api.MemberType{
			api.UserMemberType,
			api.GroupMemberType,
			api.RobotMemberType,
		}
		b := &bytes.Buffer{}
		err := json.NewEncoder(b).Encode(s)
		Expect(err).ToNot(HaveOccurred())
		Expect(b.String()).To(Equal("[\"User\",\"Group\",\"Robot\"]\n"))
	})
	It("can be json decoded", func() {
		var s []api.MemberType
		b := bytes.NewBufferString("[\"User\",\"Group\",\"Robot\"]\n")
		err := json.NewDecoder(b).Decode(&s)
		Expect(err).ToNot(HaveOccurred())
		Expect(s).To(Equal([]api.MemberType{
			api.UserMemberType,
			api.GroupMemberType,
			api.RobotMemberType,
		}))
	})

})

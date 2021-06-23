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

package harbor

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"encoding/json"
)

var _ = Describe("Role", func() {
	It("can produce valid string representation", func() {
		Expect(projectAdminRole.String()).To(Equal("ProjectAdmin"))
		Expect(developerRole.String()).To(Equal("Developer"))
		Expect(guestRole.String()).To(Equal("Guest"))
		Expect(maintainerRole.String()).To(Equal("Maintainer"))
		Expect(role(100).String()).To(Equal("*unknown-role*"))
	})
	It("can be encoded to json value", func() {
		s := struct {
			P role
			D role
			G role
			M role
		}{
			P: projectAdminRole,
			D: developerRole,
			G: guestRole,
			M: maintainerRole,
		}
		buf := bytes.NewBuffer(nil)
		Expect(json.NewEncoder(buf).Encode(&s)).To(Succeed())
		Expect(buf).To(MatchJSON(`
                      {
                        "P": 1,
                        "D": 2,
                        "G": 3,
                        "M": 4
                      }`))
	})
	It("can encode to json with unknown role", func() {
		s := struct {
			P role
		}{
			P: role(100),
		}
		buf := bytes.NewBuffer(nil)
		Expect(json.NewEncoder(buf).Encode(&s)).To(Succeed())
		Expect(buf).To(MatchJSON(`
                      {
                        "P": 100
                      }`))
	})
	It("can be decoded from json value", func() {
		buf := bytes.NewBufferString(`{
                        "P": 1,
                        "D": 2,
                        "G": 3,
                        "M": 4
                      }`)
		var roles struct {
			P role
			D role
			G role
			M role
		}
		Expect(json.NewDecoder(buf).Decode(&roles)).To(Succeed())
		Expect(roles).To(Equal(struct {
			P role
			D role
			G role
			M role
		}{
			P: projectAdminRole,
			D: developerRole,
			G: guestRole,
			M: maintainerRole,
		}))
	})
	It("fails to decode from invalid json value", func() {
		buf := bytes.NewBufferString(`{
                        "P": "ProjectAdmin"
                      }`)
		var roles struct {
			P role
		}
		Expect(json.NewDecoder(buf).Decode(&roles)).NotTo(Succeed())
	})
})

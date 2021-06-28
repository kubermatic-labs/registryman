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

var (
	proj1 = reconciler.ProjectStatus{
		Name: "proj1",
		Members: []reconciler.MemberStatus{
			reconciler.MemberStatus{
				Name: "admin",
				Type: "User",
				Role: "admin",
			},
		},
	}

	proj2 = reconciler.ProjectStatus{
		Name: "proj2",
		Members: []reconciler.MemberStatus{
			reconciler.MemberStatus{
				Name: "admin",
				Type: "User",
				Role: "admin",
			},
		},
	}

	proj1Prime = reconciler.ProjectStatus{
		Name: "proj1",
		Members: []reconciler.MemberStatus{
			reconciler.MemberStatus{
				Name: "admin",
				Type: "User",
				Role: "admin",
			},
			reconciler.MemberStatus{
				Name: "alpha",
				Type: "User",
				Role: "Developer",
			},
		},
	}

	proj1Local = reconciler.ProjectStatus{
		Name: "proj1",
		Members: []reconciler.MemberStatus{
			reconciler.MemberStatus{
				Name: "admin",
				Type: "User",
				Role: "admin",
			},
		},
	}

	bug1_actual = []reconciler.ProjectStatus{
		reconciler.ProjectStatus{
			Name: "app-images",
			Members: []reconciler.MemberStatus{
				reconciler.MemberStatus{
					Name: "alpha",
					Type: "User",
					Role: "Maintainer",
				},
				reconciler.MemberStatus{
					Name: "beta",
					Type: "User",
					Role: "Developer",
				},
			},
			ReplicationRules: []reconciler.ReplicationRuleStatus{},
		},
		reconciler.ProjectStatus{
			Name: "os-images",
			Members: []reconciler.MemberStatus{
				reconciler.MemberStatus{
					Name: "admin",
					Type: "User",
					Role: "ProjectAdmin",
				},
				reconciler.MemberStatus{
					Name: "alpha",
					Type: "User",
					Role: "Maintainer",
				},
				reconciler.MemberStatus{
					Name: "beta",
					Type: "User",
					Role: "Developer",
				},
			},
			ReplicationRules: []reconciler.ReplicationRuleStatus{},
		},
	}

	bug1_expected = []reconciler.ProjectStatus{
		reconciler.ProjectStatus{
			Name: "app-images",
			Members: []reconciler.MemberStatus{
				reconciler.MemberStatus{
					Name: "alpha",
					Type: "User",
					Role: "Maintainer",
				},
				reconciler.MemberStatus{
					Name: "beta",
					Type: "User",
					Role: "Developer",
				},
			},
			ReplicationRules: []reconciler.ReplicationRuleStatus{},
		},
	}
)

var _ = Describe("Projectstatus", func() {
	It("returns no action for the same projects", func() {
		act := []reconciler.ProjectStatus{}
		exp := []reconciler.ProjectStatus{}
		actions := reconciler.CompareProjectStatuses(nil, act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))

		act = []reconciler.ProjectStatus{
			proj1,
		}
		exp = []reconciler.ProjectStatus{
			proj1,
		}
		actions = reconciler.CompareProjectStatuses(nil, act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))

		act = []reconciler.ProjectStatus{
			proj2,
			proj1,
		}
		exp = []reconciler.ProjectStatus{
			proj1,
			proj2,
		}
		actions = reconciler.CompareProjectStatuses(nil, act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))
	})

	It("can detect missing projects", func() {
		act := []reconciler.ProjectStatus{}
		exp := []reconciler.ProjectStatus{
			proj1,
		}
		actions := reconciler.CompareProjectStatuses(nil, act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(2))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"adding project proj1",
			"adding member admin to proj1",
		}))

		act = []reconciler.ProjectStatus{
			proj2,
		}
		exp = []reconciler.ProjectStatus{
			proj1,
			proj2,
		}
		actions = reconciler.CompareProjectStatuses(nil, act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(2))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"adding project proj1",
			"adding member admin to proj1",
		}))

	})
	It("can detect surplus projects", func() {
		act := []reconciler.ProjectStatus{
			proj1,
		}
		exp := []reconciler.ProjectStatus{}
		actions := reconciler.CompareProjectStatuses(nil, act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing project proj1",
		}))

		act = []reconciler.ProjectStatus{
			proj1,
			proj2,
		}
		exp = []reconciler.ProjectStatus{
			proj2,
		}
		actions = reconciler.CompareProjectStatuses(nil, act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing project proj1",
		}))
	})
	It("can detect changed members", func() {
		act := []reconciler.ProjectStatus{
			proj1,
		}
		exp := []reconciler.ProjectStatus{
			proj1Prime,
		}
		actions := reconciler.CompareProjectStatuses(nil, act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"adding member alpha to proj1",
		}))
	})
	It("does nothing when just the project type is changed", func() {
		act := []reconciler.ProjectStatus{
			proj1,
		}
		exp := []reconciler.ProjectStatus{
			proj1Local,
		}
		actions := reconciler.CompareProjectStatuses(nil, act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))
		Expect(actionsToStrings(actions)).To(Equal([]string{}))

	})
	It("removes extra project and updates members in a single run", func() {
		act := []reconciler.ProjectStatus{
			proj1Prime,
			proj2,
		}
		exp := []reconciler.ProjectStatus{
			proj1,
		}
		actions := reconciler.CompareProjectStatuses(nil, act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(2))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing project proj2",
			"removing member alpha from proj1",
		}))
	})
	It("acts as expected for bug1", func() {
		act := bug1_actual
		exp := bug1_expected
		actions := reconciler.CompareProjectStatuses(nil, act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing project os-images",
		}))
	})
})

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

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry/reconciler"
)

var (
	proj1 = api.ProjectStatus{
		Name: "proj1",
		Members: []api.MemberStatus{
			api.MemberStatus{
				Name: "admin",
				Type: "User",
				Role: "admin",
			},
		},
	}

	proj2 = api.ProjectStatus{
		Name: "proj2",
		Members: []api.MemberStatus{
			api.MemberStatus{
				Name: "admin",
				Type: "User",
				Role: "admin",
			},
		},
	}

	proj1Prime = api.ProjectStatus{
		Name: "proj1",
		Members: []api.MemberStatus{
			api.MemberStatus{
				Name: "admin",
				Type: "User",
				Role: "admin",
			},
			api.MemberStatus{
				Name: "alpha",
				Type: "User",
				Role: "Developer",
			},
		},
	}

	proj1Local = api.ProjectStatus{
		Name: "proj1",
		Members: []api.MemberStatus{
			api.MemberStatus{
				Name: "admin",
				Type: "User",
				Role: "admin",
			},
		},
	}

	bug1_actual = []api.ProjectStatus{
		api.ProjectStatus{
			Name: "app-images",
			Members: []api.MemberStatus{
				api.MemberStatus{
					Name: "alpha",
					Type: "User",
					Role: "Maintainer",
				},
				api.MemberStatus{
					Name: "beta",
					Type: "User",
					Role: "Developer",
				},
			},
			ReplicationRules: []api.ReplicationRuleStatus{},
		},
		api.ProjectStatus{
			Name: "os-images",
			Members: []api.MemberStatus{
				api.MemberStatus{
					Name: "admin",
					Type: "User",
					Role: "ProjectAdmin",
				},
				api.MemberStatus{
					Name: "alpha",
					Type: "User",
					Role: "Maintainer",
				},
				api.MemberStatus{
					Name: "beta",
					Type: "User",
					Role: "Developer",
				},
			},
			ReplicationRules: []api.ReplicationRuleStatus{},
		},
	}

	bug1_expected = []api.ProjectStatus{
		api.ProjectStatus{
			Name: "app-images",
			Members: []api.MemberStatus{
				api.MemberStatus{
					Name: "alpha",
					Type: "User",
					Role: "Maintainer",
				},
				api.MemberStatus{
					Name: "beta",
					Type: "User",
					Role: "Developer",
				},
			},
			ReplicationRules: []api.ReplicationRuleStatus{},
		},
	}
)

var _ = Describe("Projectstatus", func() {
	It("returns no action for the same projects", func() {
		act := []api.ProjectStatus{}
		exp := []api.ProjectStatus{}
		actions := reconciler.CompareProjectStatuses(nil, act, exp, api.RegistryCapabilities{})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))

		act = []api.ProjectStatus{
			proj1,
		}
		exp = []api.ProjectStatus{
			proj1,
		}
		actions = reconciler.CompareProjectStatuses(nil, act, exp, api.RegistryCapabilities{})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))

		act = []api.ProjectStatus{
			proj2,
			proj1,
		}
		exp = []api.ProjectStatus{
			proj1,
			proj2,
		}
		actions = reconciler.CompareProjectStatuses(nil, act, exp, api.RegistryCapabilities{})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))
	})

	It("can detect missing projects", func() {
		act := []api.ProjectStatus{}
		exp := []api.ProjectStatus{
			proj1,
		}
		actions := reconciler.CompareProjectStatuses(nil, act, exp, api.RegistryCapabilities{
			CanCreateProject:            true,
			CanManipulateProjectMembers: true,
		})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(2))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"adding project proj1",
			"adding member admin to proj1",
		}))

		act = []api.ProjectStatus{
			proj2,
		}
		exp = []api.ProjectStatus{
			proj1,
			proj2,
		}
		actions = reconciler.CompareProjectStatuses(nil, act, exp, api.RegistryCapabilities{
			CanCreateProject:            true,
			CanManipulateProjectMembers: true,
		})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(2))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"adding project proj1",
			"adding member admin to proj1",
		}))

	})
	It("can detect surplus projects", func() {
		act := []api.ProjectStatus{
			proj1,
		}
		exp := []api.ProjectStatus{}
		actions := reconciler.CompareProjectStatuses(nil, act, exp, api.RegistryCapabilities{
			CanDeleteProject: true,
		})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing project proj1",
		}))

		act = []api.ProjectStatus{
			proj1,
			proj2,
		}
		exp = []api.ProjectStatus{
			proj2,
		}
		actions = reconciler.CompareProjectStatuses(nil, act, exp, api.RegistryCapabilities{
			CanDeleteProject: true,
		})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing project proj1",
		}))
	})
	It("can detect changed members", func() {
		act := []api.ProjectStatus{
			proj1,
		}
		exp := []api.ProjectStatus{
			proj1Prime,
		}
		actions := reconciler.CompareProjectStatuses(nil, act, exp, api.RegistryCapabilities{
			CanManipulateProjectMembers: true,
		})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"adding member alpha to proj1",
		}))
	})
	It("does nothing when just the project type is changed", func() {
		act := []api.ProjectStatus{
			proj1,
		}
		exp := []api.ProjectStatus{
			proj1Local,
		}
		actions := reconciler.CompareProjectStatuses(nil, act, exp, api.RegistryCapabilities{})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))
		Expect(actionsToStrings(actions)).To(Equal([]string{}))

	})
	It("removes extra project and updates members in a single run", func() {
		act := []api.ProjectStatus{
			proj1Prime,
			proj2,
		}
		exp := []api.ProjectStatus{
			proj1,
		}
		actions := reconciler.CompareProjectStatuses(nil, act, exp, api.RegistryCapabilities{
			CanDeleteProject:            true,
			CanManipulateProjectMembers: true,
		})
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
		actions := reconciler.CompareProjectStatuses(nil, act, exp, api.RegistryCapabilities{
			CanDeleteProject: true,
		})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing project os-images",
		}))
	})
})

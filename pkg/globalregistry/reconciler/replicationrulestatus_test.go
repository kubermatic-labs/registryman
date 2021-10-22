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
	rrule1 = api.ReplicationRuleStatus{
		RemoteRegistryName: "reg1",
		Trigger: api.ReplicationTrigger{
			Type:     api.EventBasedReplicationTriggerType,
			Schedule: "",
		},
		Direction: "Push",
	}
	rrule2 = api.ReplicationRuleStatus{
		RemoteRegistryName: "reg2",
		Trigger: api.ReplicationTrigger{
			Type:     api.EventBasedReplicationTriggerType,
			Schedule: "",
		},
		Direction: "Push",
	}
	rrule1Trig = api.ReplicationRuleStatus{
		RemoteRegistryName: "reg1",
		Trigger: api.ReplicationTrigger{
			Type:     api.ManualReplicationTriggerType,
			Schedule: "",
		},
		Direction: "Push",
	}
	rrule1Pull = api.ReplicationRuleStatus{
		RemoteRegistryName: "reg1",
		Trigger: api.ReplicationTrigger{
			Type:     api.EventBasedReplicationTriggerType,
			Schedule: "",
		},
		Direction: "Pull",
	}
)

var _ = Describe("Memberstatus", func() {
	It("returns no action for the same ReplicationRuleStatus slice", func() {
		act := []api.ReplicationRuleStatus{}
		exp := []api.ReplicationRuleStatus{}
		actions := reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp, api.RegistryCapabilities{})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))

		act = []api.ReplicationRuleStatus{
			rrule1,
		}
		exp = []api.ReplicationRuleStatus{
			rrule1,
		}
		actions = reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp, api.RegistryCapabilities{})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))

		act = []api.ReplicationRuleStatus{
			rrule1,
			rrule2,
		}
		exp = []api.ReplicationRuleStatus{
			rrule1,
			rrule2,
		}
		actions = reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp, api.RegistryCapabilities{})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))

		act = []api.ReplicationRuleStatus{
			rrule1,
			rrule2,
		}
		exp = []api.ReplicationRuleStatus{
			rrule2,
			rrule1,
		}
		actions = reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp, api.RegistryCapabilities{})
		Expect(actions).ToNot(BeNil())
	})

	It("can detect missing rules", func() {
		act := []api.ReplicationRuleStatus{}
		exp := []api.ReplicationRuleStatus{
			rrule1,
		}
		actions := reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp, api.RegistryCapabilities{
			CanManipulateProjectReplicationRules: true,
		})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"adding replication rule for proj: reg1 [Push] on event_based",
		}))

		act = []api.ReplicationRuleStatus{
			rrule2,
		}
		exp = []api.ReplicationRuleStatus{
			rrule1,
			rrule2,
		}
		actions = reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp, api.RegistryCapabilities{
			CanManipulateProjectReplicationRules: true,
		})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"adding replication rule for proj: reg1 [Push] on event_based",
		}))
	})

	It("can detect surplus rules", func() {
		act := []api.ReplicationRuleStatus{
			rrule1,
		}
		exp := []api.ReplicationRuleStatus{}
		actions := reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp, api.RegistryCapabilities{
			CanManipulateProjectReplicationRules: true,
		})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing replication rule for proj: reg1 [Push] on event_based",
		}))
		act = []api.ReplicationRuleStatus{
			rrule1,
			rrule2,
		}
		exp = []api.ReplicationRuleStatus{
			rrule2,
		}
		actions = reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp, api.RegistryCapabilities{
			CanManipulateProjectReplicationRules: true,
		})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing replication rule for proj: reg1 [Push] on event_based",
		}))
	})

	It("can detect different rules", func() {
		act := []api.ReplicationRuleStatus{
			rrule1,
		}
		exp := []api.ReplicationRuleStatus{
			rrule2,
		}
		actions := reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp, api.RegistryCapabilities{
			CanManipulateProjectReplicationRules: true,
		})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(2))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing replication rule for proj: reg1 [Push] on event_based",
			"adding replication rule for proj: reg2 [Push] on event_based",
		}))
		act = []api.ReplicationRuleStatus{
			rrule1,
		}
		exp = []api.ReplicationRuleStatus{
			rrule1Trig,
		}
		actions = reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp, api.RegistryCapabilities{
			CanManipulateProjectReplicationRules: true,
		})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(2))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing replication rule for proj: reg1 [Push] on event_based",
			"adding replication rule for proj: reg1 [Push] on manual",
		}))
		act = []api.ReplicationRuleStatus{
			rrule1,
		}
		exp = []api.ReplicationRuleStatus{
			rrule1Pull,
		}
		actions = reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp, api.RegistryCapabilities{
			CanManipulateProjectReplicationRules: true,
		})
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(2))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing replication rule for proj: reg1 [Push] on event_based",
			"adding replication rule for proj: reg1 [Pull] on event_based",
		}))
	})
})

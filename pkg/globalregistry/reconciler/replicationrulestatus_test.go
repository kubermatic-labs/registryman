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

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry/reconciler"
)

var (
	rrule1 = reconciler.ReplicationRuleStatus{
		RemoteRegistryName: "reg1",
		Trigger:            globalregistry.EventReplicationTrigger,
		Direction:          globalregistry.PushReplication,
	}
	rrule2 = reconciler.ReplicationRuleStatus{
		RemoteRegistryName: "reg2",
		Trigger:            globalregistry.EventReplicationTrigger,
		Direction:          globalregistry.PushReplication,
	}
	rrule1Trig = reconciler.ReplicationRuleStatus{
		RemoteRegistryName: "reg1",
		Trigger:            globalregistry.ManualReplicationTrigger,
		Direction:          globalregistry.PushReplication,
	}
	rrule1Pull = reconciler.ReplicationRuleStatus{
		RemoteRegistryName: "reg1",
		Trigger:            globalregistry.EventReplicationTrigger,
		Direction:          globalregistry.PullReplication,
	}
)

var _ = Describe("Memberstatus", func() {
	It("returns no action for the same ReplicationRuleStatus slice", func() {
		act := []reconciler.ReplicationRuleStatus{}
		exp := []reconciler.ReplicationRuleStatus{}
		actions := reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))

		act = []reconciler.ReplicationRuleStatus{
			rrule1,
		}
		exp = []reconciler.ReplicationRuleStatus{
			rrule1,
		}
		actions = reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))

		act = []reconciler.ReplicationRuleStatus{
			rrule1,
			rrule2,
		}
		exp = []reconciler.ReplicationRuleStatus{
			rrule1,
			rrule2,
		}
		actions = reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(0))

		act = []reconciler.ReplicationRuleStatus{
			rrule1,
			rrule2,
		}
		exp = []reconciler.ReplicationRuleStatus{
			rrule2,
			rrule1,
		}
		actions = reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp)
		Expect(actions).ToNot(BeNil())
	})

	It("can detect missing rules", func() {
		act := []reconciler.ReplicationRuleStatus{}
		exp := []reconciler.ReplicationRuleStatus{
			rrule1,
		}
		actions := reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"adding replication rule for proj: reg1 [Push] on EventBased",
		}))

		act = []reconciler.ReplicationRuleStatus{
			rrule2,
		}
		exp = []reconciler.ReplicationRuleStatus{
			rrule1,
			rrule2,
		}
		actions = reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"adding replication rule for proj: reg1 [Push] on EventBased",
		}))
	})

	It("can detect surplus rules", func() {
		act := []reconciler.ReplicationRuleStatus{
			rrule1,
		}
		exp := []reconciler.ReplicationRuleStatus{}
		actions := reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing replication rule for proj: reg1 [Push] on EventBased",
		}))
		act = []reconciler.ReplicationRuleStatus{
			rrule1,
			rrule2,
		}
		exp = []reconciler.ReplicationRuleStatus{
			rrule2,
		}
		actions = reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(1))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing replication rule for proj: reg1 [Push] on EventBased",
		}))
	})

	It("can detect different rules", func() {
		act := []reconciler.ReplicationRuleStatus{
			rrule1,
		}
		exp := []reconciler.ReplicationRuleStatus{
			rrule2,
		}
		actions := reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(2))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing replication rule for proj: reg1 [Push] on EventBased",
			"adding replication rule for proj: reg2 [Push] on EventBased",
		}))
		act = []reconciler.ReplicationRuleStatus{
			rrule1,
		}
		exp = []reconciler.ReplicationRuleStatus{
			rrule1Trig,
		}
		actions = reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(2))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing replication rule for proj: reg1 [Push] on EventBased",
			"adding replication rule for proj: reg1 [Push] on Manual",
		}))
		act = []reconciler.ReplicationRuleStatus{
			rrule1,
		}
		exp = []reconciler.ReplicationRuleStatus{
			rrule1Pull,
		}
		actions = reconciler.CompareReplicationRuleStatus(nil, "proj", act, exp)
		Expect(actions).ToNot(BeNil())
		Expect(len(actions)).To(Equal(2))
		Expect(actionsToStrings(actions)).To(Equal([]string{
			"removing replication rule for proj: reg1 [Push] on EventBased",
			"adding replication rule for proj: reg1 [Pull] on EventBased",
		}))
	})
})

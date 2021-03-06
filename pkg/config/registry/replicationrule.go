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

package registry

import (
	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type calculatedReplication int

const (
	noReplication calculatedReplication = iota
	pullReplication
	pushReplication
)

type registryCapabilities struct {
	isGlobal bool
	globalregistry.ReplicationCapabilities
}

func calculateReplicationRule(local, remote registryCapabilities) calculatedReplication {
	if local.isGlobal && remote.isGlobal {
		panic("both local and remote are global")
	}
	if !local.isGlobal && !remote.isGlobal {
		return noReplication
	}
	if remote.isGlobal && remote.CanPush() {
		return noReplication
	}
	if local.isGlobal && local.CanPush() {
		return pushReplication
	}
	if !local.isGlobal && local.CanPull() {
		return pullReplication
	}
	return noReplication
}

type replicationRule struct {
	calculatedReplication
	project *project
	remote  *Registry
}

var _ globalregistry.ReplicationRule = &replicationRule{}

func (rule *replicationRule) GetProjectName() string {
	return rule.project.GetName()
}

func (rule *replicationRule) GetName() string {
	panic("not implemented")
}

type replicationTrigger struct {
	triggerType     api.ReplicationTriggerType
	triggerSchedule string
}

var _ globalregistry.ReplicationTrigger = replicationTrigger{}

func (rt replicationTrigger) TriggerType() api.ReplicationTriggerType {
	return rt.triggerType
}

func (rt replicationTrigger) TriggerSchedule() string {
	return rt.triggerSchedule
}

var (
	eventBasedReplicationTrigger = replicationTrigger{api.EventBasedReplicationTriggerType, ""}
	fallbackTrigger              = replicationTrigger{api.CronReplicationTriggerType, "*/10 * * * *"}
)

func (rule *replicationRule) Trigger() globalregistry.ReplicationTrigger {
	switch rule.calculatedReplication {
	case noReplication:
		panic("noReplication not handled")

		// In case of push replication we respect cron and manual,
		// otherwise use the event-based replication trigger
	case pushReplication:
		switch rule.project.Spec.Trigger.Type {
		case api.CronReplicationTriggerType, api.ManualReplicationTriggerType:
			return rule.project.Spec.Trigger
		default:
			return eventBasedReplicationTrigger
		}

		// In case of pull replication we respect cron and manual,
		// otherwise use the fallback trigger
	case pullReplication:
		switch rule.project.Spec.Trigger.Type {
		case api.CronReplicationTriggerType, api.ManualReplicationTriggerType:
			return rule.project.Spec.Trigger
		default:
			return fallbackTrigger
		}
	default:
		panic("unhandled case")
	}

}

func (rule *replicationRule) Direction() string {
	switch rule.calculatedReplication {
	case noReplication:
		panic("noReplication not handled")
	case pushReplication:
		return "Push"
	case pullReplication:
		return "Pull"
	default:
		panic("unhandled case")
	}

}

func (rule *replicationRule) RemoteRegistry() globalregistry.Registry {
	return rule.remote
}

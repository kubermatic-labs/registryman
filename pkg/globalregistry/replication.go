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
package globalregistry

import (
	"fmt"
)

type ReplicationDirection int

const (
	PullReplication ReplicationDirection = iota
	PushReplication
)

func (rt ReplicationDirection) String() string {
	switch rt {
	case PullReplication:
		return "Pull"
	case PushReplication:
		return "Push"
	default:
		return fmt.Sprintf("*unknown ReplicationType: %d*", int(rt))
	}
}

type ReplicationTrigger int

const (
	ManualReplicationTrigger ReplicationTrigger = iota
	EventReplicationTrigger
)

func (rt ReplicationTrigger) String() string {
	switch rt {
	case ManualReplicationTrigger:
		return "Manual"
	case EventReplicationTrigger:
		return "EventBased"
	default:
		panic("unhandled ReplicationTrigger value")
	}
}

type ReplicationRule interface {
	GetProjectName() string
	GetName() string
	Trigger() ReplicationTrigger
	Direction() ReplicationDirection
	RemoteRegistry() Registry
	Delete() error
}

type ReplicationAPI interface {
	List() ([]ReplicationRule, error)
}

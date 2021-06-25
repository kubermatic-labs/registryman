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

// ReplicationDirection shows the Project replication direction. In case of
// PullReplication, the registry of the project will pull the repositories from
// a remote registry. In case of PushReplication, the registry will push the
// repos.
type ReplicationDirection int

const (
	PullReplication ReplicationDirection = iota
	PushReplication
)

func (rd ReplicationDirection) string() (string, error) {
	switch rd {
	case PullReplication:
		return "Pull", nil
	case PushReplication:
		return "Push", nil
	default:
		return "", fmt.Errorf("unknown ReplicationType: %d", int(rd))
	}
}

func (rd ReplicationDirection) String() string {
	s, err := rd.string()
	if err != nil {
		panic(err.Error())
	}
	return s
}

// MarshalText method implements the encoding.TextMarshaler interface.
func (rd ReplicationDirection) MarshalText() ([]byte, error) {
	s, err := rd.string()
	return []byte(s), err
}

// ReplicationTrigger describes the trigger event that starts the
// synchronization mechanism of the project.
type ReplicationTrigger int

const (
	ManualReplicationTrigger ReplicationTrigger = iota
	EventReplicationTrigger
)

func (rt ReplicationTrigger) string() (string, error) {
	switch rt {
	case ManualReplicationTrigger:
		return "Manual", nil
	case EventReplicationTrigger:
		return "EventBased", nil
	default:
		return "", fmt.Errorf("unhandled ReplicationTrigger value: %d", rt)
	}
}

func (rt ReplicationTrigger) String() string {
	s, err := rt.string()
	if err != nil {
		panic(err.Error())
	}
	return s
}

// MarshalText method implements the encoding.TextMarshaler interface.
func (rt ReplicationTrigger) MarshalText() ([]byte, error) {
	s, err := rt.string()
	return []byte(s), err
}

// ReplicationRule interface declares the methods that can be used to manipulate
// the replication rule of a project.
type ReplicationRule interface {

	// GetProjectName returns the name of the project that is subject to the
	// replication.
	GetProjectName() string

	// GetName returns the name of the replication rule.
	GetName() string

	// Trigger returns the event that starts the replication.
	Trigger() ReplicationTrigger

	// Direction returns the direction of the synchronization.
	Direction() ReplicationDirection

	// RemoteRegistry returns the remote registry which is subject to the
	// replication.
	RemoteRegistry() Registry

	// Delete method deletes the replication rule from the registry.
	Delete() error
}

// ReplicationAPI interface defines the methods of a registry which are related
// to the management of the replication rules.
type ReplicationAPI interface {
	// List method returns the replication rules of a registry.
	List() ([]ReplicationRule, error)
}

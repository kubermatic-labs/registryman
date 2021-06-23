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

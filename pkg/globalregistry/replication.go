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

import "context"

// ReplicationRule interface declares the methods that can be used to manipulate
// the replication rule of a project.
type ReplicationRule interface {

	// GetProjectName returns the name of the project that is subject to the
	// replication.
	GetProjectName() string

	// GetName returns the name of the replication rule.
	GetName() string

	// Trigger returns the event that starts the replication.
	Trigger() string

	// Direction returns the direction of the synchronization.
	Direction() string

	// RemoteRegistry returns the remote registry which is subject to the
	// replication.
	RemoteRegistry() Registry
}

// DestructibleReplicationRule interface declares the methods that can be used
// to delete the replication rule of a project.
type DestructibleReplicationRule interface {
	// Delete method deletes the replication rule from the registry.
	Delete(context.Context) error
}

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
	"context"
	"fmt"
	"strings"

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type replicationFilter struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type triggerSettings struct {
	Cron string `json:"cron"`
}

type replicationTrigger struct {
	Type            string          `json:"type"`
	TriggerSettings triggerSettings `json:"trigger_settings"`
}

func (rt replicationTrigger) TriggerType() api.ReplicationTriggerType {
	var tt api.ReplicationTriggerType
	err := tt.UnmarshalText([]byte(rt.Type))
	if err != nil {
		panic(fmt.Errorf("unknown trigger type: %s", rt.Type))
	}
	return tt
}

func (rt replicationTrigger) TriggerSchedule() string {
	scheduleWords := strings.SplitN(rt.TriggerSettings.Cron, "", 2)
	if len(scheduleWords) != 2 {
		return rt.TriggerSettings.Cron
	}
	return scheduleWords[1]
}

type replicationResponseBody struct {
	UpdateTime    string                `json:"update_time"`
	Description   string                `json:"description"`
	Enabled       bool                  `json:"enabled"`
	Filters       []replicationFilter   `json:"filters"`
	DestRegistry  *remoteRegistryStatus `json:"dest_registry,omitempty"`
	CreationTime  string                `json:"creation_time"`
	SrcRegistry   *remoteRegistryStatus `json:"src_registry,omitempty"`
	DestNamespace string                `json:"dest_namespace"`
	Trigger       *replicationTrigger   `json:"trigger"`
	Deletion      bool                  `json:"deletion"`
	Override      bool                  `json:"override"`
	Id            int                   `json:"id"`
	Name          string                `json:"name"`
}

func (rp *replicationResponseBody) direction() (string, error) {
	if rp.SrcRegistry.Name == "Local" {
		return "Push", nil
	}
	if rp.DestRegistry.Name == "Local" {
		return "Pull", nil
	}
	return "Pull", fmt.Errorf("cannot determine direction")
}

func (rp *replicationResponseBody) remote() (*remoteRegistryStatus, error) {
	if rp.SrcRegistry.Name == "Local" {
		return rp.DestRegistry, nil
	}
	if rp.DestRegistry.Name == "Local" {
		return rp.SrcRegistry, nil
	}
	return nil, fmt.Errorf("cannot determine direction")
}

//replicationRule implements the globalregistry.replicationRule.
type replicationRule struct {
	ID          int
	registry    *registry
	name        string
	projectName string
	Dir         string
	ReplTrigger *replicationTrigger
	Remote      *remoteRegistryStatus
}

var _ globalregistry.ReplicationRule = &replicationRule{}
var _ globalregistry.DestructibleReplicationRule = &replicationRule{}
var _ globalregistry.UpdatableRemoteRegistryReplicationRule = &replicationRule{}

func (r *replicationRule) GetProjectName() string {
	return r.projectName
}

func (r *replicationRule) GetName() string {
	return r.name
}

func (r *replicationRule) Trigger() globalregistry.ReplicationTrigger {
	return r.ReplTrigger
}

func (r *replicationRule) Direction() string {
	return r.Dir
}

func (r *replicationRule) RemoteRegistry() globalregistry.Registry {
	return r.Remote
}

func (r *replicationRule) UpdateRemoteRegistry(ctx context.Context, remoteRegistry globalregistry.Registry) error {
	return r.registry.updateRemoteRegistry(ctx, r.Remote.Id, remoteRegistry)
}

func (r *replicationRule) Delete(ctx context.Context) error {
	return r.registry.deleteReplicationRule(ctx, r.ID)
}

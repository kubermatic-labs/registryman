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
	"context"
	"fmt"

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type project struct {
	*api.Project
	registry *Registry
}

var _ globalregistry.Project = &project{}
var _ globalregistry.ProjectWithMembers = &project{}
var _ globalregistry.ProjectWithReplication = &project{}
var _ globalregistry.ProjectWithScanner = &project{}

func (proj *project) GetMembers(context.Context) ([]globalregistry.ProjectMember, error) {
	members := make([]globalregistry.ProjectMember, len(proj.Spec.Members))
	for i, member := range proj.Spec.Members {
		pMember := &projectMember{
			ProjectMember: member,
		}
		if member.DN != "" {
			fmt.Println("GROUP member")
			members[i] = &ldapGroupMember{
				pMember,
			}
		} else {
			members[i] = pMember
		}
	}
	return members, nil
}

func (proj *project) GetReplicationRules(ctx context.Context, trigger, direction string) ([]globalregistry.ReplicationRule, error) {
	rules := []globalregistry.ReplicationRule{}
	switch proj.Spec.Type {
	case api.GlobalProjectType:
		for _, r := range proj.registry.apiProvider.GetRegistries(ctx) {
			remoteReg := New(r, proj.registry.apiProvider)
			if proj.registry.GetName() != r.GetName() {
				calcRepl := calculateReplicationRule(
					proj.registry.registryCapabilities(),
					remoteReg.registryCapabilities(),
				)
				if calcRepl != noReplication {
					repRule := &replicationRule{
						calculatedReplication: calcRepl,
						project:               proj,
						remote:                remoteReg,
					}
					if trigger != "" && trigger != repRule.Trigger() {
						continue
					}
					if direction != "" && direction != repRule.Direction() {
						continue
					}
					rules = append(rules, repRule)
				}
			}
		}
	case api.LocalProjectType:
	default:
		return nil, fmt.Errorf("invalid registry type: %s", proj.Spec.Type.String())
	}
	return rules, nil
}

func (p *project) GetScanner(ctx context.Context) (globalregistry.Scanner, error) {
	if p.Spec.Scanner == "" {
		return nil, nil
	}
	scanners := p.registry.apiProvider.GetScanners(ctx)
	for _, s := range scanners {
		if s.GetName() == p.Spec.Scanner {
			return &scanner{s}, nil
		}
	}
	return nil, fmt.Errorf("project %s has invalid scanner configuration (%s)", p.GetName(), p.Spec.Scanner)
}

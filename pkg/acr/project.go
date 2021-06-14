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
package acr

import (
	"fmt"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type project struct {
	name string
	api  *projectAPI
}

func (p *project) GetName() string {
	return p.name
}

func (p *project) Delete() error {
	return fmt.Errorf("project deletion on ACR is not implemented: %w", globalregistry.RecoverableError)
	//	return p.api.delete(p.Name)
}

func (p *project) AssignMember(member globalregistry.ProjectMember) (*globalregistry.ProjectMemberCredentials, error) {
	return nil, fmt.Errorf("method ACR.AssignMember not implemented: %w", globalregistry.RecoverableError)
}

func (p *project) UnassignMember(member globalregistry.ProjectMember) error {
	return globalregistry.RecoverableError
}

func (p *project) AssignReplicationRule(remoteReg globalregistry.RegistryConfig, trigger globalregistry.ReplicationTrigger, direction globalregistry.ReplicationDirection) (globalregistry.ReplicationRule, error) {
	return nil, globalregistry.RecoverableError
}

func (p *project) GetMembers() ([]globalregistry.ProjectMember, error) {
	p.api.reg.logger.V(-1).Info("ACR.GetMembers not implemented")
	return []globalregistry.ProjectMember{}, nil
}

func (p *project) GetReplicationRules(
	trigger *globalregistry.ReplicationTrigger,
	direction *globalregistry.ReplicationDirection) ([]globalregistry.ReplicationRule, error) {

	return nil, nil
}

func (p *project) AssignScanner(s globalregistry.Scanner) error {
	return fmt.Errorf("method ACR.AssignScanner not implemented: %w", globalregistry.RecoverableError)
}

func (p *project) GetScanner() (globalregistry.Scanner, error) {
	return nil, nil
}

func (p *project) UnassignScanner(s globalregistry.Scanner) error {
	return fmt.Errorf("method ACR.UnassignScanner not implemented: %w", globalregistry.RecoverableError)
}

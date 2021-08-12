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

var _ globalregistry.RegistryWithProjects = &Registry{}

func (r *Registry) GetProjectByName(name string) (globalregistry.Project, error) {
	if name == "" {
		return &project{
			Project:  nil,
			registry: r,
		}, nil
	}
	projects := r.apiProvider.GetProjects()
	for _, proj := range projects {
		if proj.GetName() == name {
			return &project{
				Project:  proj,
				registry: r,
			}, nil
		}
	}
	return nil, nil
}

func (r *Registry) ListProjects() ([]globalregistry.Project, error) {
	projects := r.apiProvider.GetProjects()
	result := make([]globalregistry.Project, 0)
	for _, proj := range projects {
		myProject := false
	LRegLoop:
		for _, lReg := range proj.Spec.LocalRegistries {
			if lReg == r.GetName() {
				myProject = true
				break LRegLoop
			}
		}
		if proj.Spec.Type == api.GlobalProjectType ||
			myProject {
			result = append(result, &project{
				Project:  proj,
				registry: r,
			})
		}
	}
	return result, nil
}

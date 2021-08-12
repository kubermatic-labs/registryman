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

package config

import (
	"fmt"
	"net/url"

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/config/registry"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

// ProjectOfRegistry struct describes a connection between a
// globalregistry.Registry and globalregistry.Project object.
type ProjectOfRegistry struct {
	Registry globalregistry.Registry
	Project  globalregistry.Project
}

// GenerateProjectRepoName generates a project-level repository URL for a given
// project.
func (p *ProjectOfRegistry) GenerateProjectRepoName() (string, error) {
	url, err := url.Parse(p.Registry.GetAPIEndpoint())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", url.Host, p.Project.GetName()), nil
}

func newProject(aos ApiObjectStore, reg *api.Registry, proj *api.Project) (*ProjectOfRegistry, error) {
	realRegistry, err := registry.New(reg, aos).ToReal()
	if err != nil {
		return nil, err
	}
	realRegistryWithProjects := realRegistry.(globalregistry.RegistryWithProjects)
	realProject, err := realRegistryWithProjects.GetProjectByName(proj.GetName())
	if err != nil {
		return nil, err
	}
	if realProject == nil {
		return nil, fmt.Errorf("%v Project doesn't exists in %v Registry, actual state differs from the expected state",
			proj.GetName(), realRegistry.GetAPIEndpoint())
	}
	return &ProjectOfRegistry{Registry: realRegistry, Project: realProject}, nil
}

// GetProjectByName returns a Project struct for a matching project name.
func GetProjectByName(aos ApiObjectStore, projectName string) (*ProjectOfRegistry, error) {
	projects := aos.GetProjects()
	registries := aos.GetRegistries()
	for _, project := range projects {
		if project.GetName() == projectName {
			switch project.Spec.Type {
			case api.GlobalProjectType:
				for _, reg := range registries {
					if reg.Spec.Role == "GlobalHub" {
						return newProject(aos, reg, project)
					}
				}
			case api.LocalProjectType:
				if len(project.Spec.LocalRegistries) == 0 {
					return nil, fmt.Errorf("local project with no local registries")
				}
				localRegistryName := project.Spec.LocalRegistries[0]
				for _, reg := range registries {
					if reg.Spec.Role == "Local" && reg.GetName() == localRegistryName {
						return newProject(aos, reg, project)
					}
				}
			default:
				panic(fmt.Sprintf("unhandled project type: %s", project.Spec.Type.String()))
			}
		}
	}
	return nil, fmt.Errorf("project %s not found", projectName)
}

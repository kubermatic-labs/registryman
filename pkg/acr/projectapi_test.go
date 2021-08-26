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
	"testing"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

func sliceContainsString(projects []globalregistry.Project, s string) bool {
	for _, proj := range projects {
		if proj.GetName() == s {
			return true
		}
	}
	return false
}

func TestCollectProjectNamesFromRepos(t *testing.T) {
	reg := &registry{}
	repos := []string{"os-images/ubuntu"}
	projects := reg.collectProjectNamesFromRepos(repos)
	if projLen := len(projects); projLen != 1 {
		t.Errorf("len of projects is %d", projLen)
	}
	if projName := projects[0].GetName(); projName != "os-images" {
		t.Errorf("invalid project name: %s", projName)
	}

	repos = []string{"os-images/ubuntu", "os-images/alpine"}
	projects = reg.collectProjectNamesFromRepos(repos)
	if projLen := len(projects); projLen != 1 {
		t.Errorf("len of projects is %d", projLen)
	}
	if projName := projects[0].GetName(); projName != "os-images" {
		t.Errorf("invalid project name: %s", projName)
	}

	repos = []string{"os-images/ubuntu", "os-images/alpine", "app-images/service"}
	projects = reg.collectProjectNamesFromRepos(repos)
	if projLen := len(projects); projLen != 2 {
		t.Errorf("len of projects is %d", projLen)
	}

	if !sliceContainsString(projects, "os-images") {
		t.Errorf("project name missing: os-images")
	}
	if !sliceContainsString(projects, "app-images") {
		t.Errorf("project name missing: app-images")
	}

	repos = []string{"ubuntu", "os-images/alpine", "app-images/service"}
	projects = reg.collectProjectNamesFromRepos(repos)
	if projLen := len(projects); projLen != 3 {
		t.Errorf("len of projects is %d", projLen)
	}

	if !sliceContainsString(projects, "os-images") {
		t.Errorf("project name missing: os-images")
	}
	if !sliceContainsString(projects, "app-images") {
		t.Errorf("project name missing: app-images")
	}
	if !sliceContainsString(projects, "ubuntu") {
		t.Errorf("project name missing: ubuntu")
	}

	repos = []string{}
	projects = reg.collectProjectNamesFromRepos(repos)
	if projLen := len(projects); projLen != 0 {
		t.Errorf("len of projects is %d", projLen)
	}
}

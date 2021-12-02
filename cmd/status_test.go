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

package cmd

import "testing"

func TestRegistryInScope(t *testing.T) {
	filteredRegistries = []string{}
	if !registryInScope("") {
		t.Errorf("registryInScope shall return true when filteredRegistries is empty")
	}
	if !registryInScope("test") {
		t.Errorf("registryInScope shall return true when filteredRegistries is empty")
	}
	filteredRegistries = []string{"reg1", "reg2", "reg3"}
	if !registryInScope("reg1") {
		t.Errorf("registryInScope shall return true for reg1")
	}
	if !registryInScope("reg2") {
		t.Errorf("registryInScope shall return true for reg2")
	}
	if !registryInScope("reg3") {
		t.Errorf("registryInScope shall return true for reg3")
	}
	if registryInScope("reg4") {
		t.Errorf("registryInScope shall return false for reg4")
	}
}

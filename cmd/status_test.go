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

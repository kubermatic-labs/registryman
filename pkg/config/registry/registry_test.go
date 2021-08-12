package registry

import (
	"fmt"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockApiProvider struct {
	forceDelete bool
}

func (ap *mockApiProvider) GetProjects() []*api.Project                              { return nil }
func (ap *mockApiProvider) GetRegistries() []*api.Registry                           { return nil }
func (ap *mockApiProvider) GetScanners() []*api.Scanner                              { return nil }
func (ap *mockApiProvider) GetGlobalRegistryOptions() globalregistry.RegistryOptions { return ap }
func (ap *mockApiProvider) GetLogger() logr.Logger                                   { return logger }
func (ap *mockApiProvider) ForceDeleteProjects() bool                                { return ap.forceDelete }

var _ ApiObjectProvider = &mockApiProvider{}

var (
	logger                 = zapr.NewLogger(zap.NewNop())
	apiProviderForceDelete = &mockApiProvider{
		forceDelete: true,
	}
	apiProviderNoForceDelete = &mockApiProvider{
		forceDelete: false,
	}
	registryForceDelete = &api.Registry{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				"registryman.kubermatic.com/forceDelete": "true",
			},
		},
	}
	registryNoForceDelete = &api.Registry{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				"registryman.kubermatic.com/forceDelete": "false",
			},
		},
	}
	registryInvalidForceDelete = &api.Registry{
		ObjectMeta: v1.ObjectMeta{
			Annotations: map[string]string{
				"registryman.kubermatic.com/forceDelete": "invalid",
			},
		},
	}
	registryMissingForceDelete = &api.Registry{}
)

func canForceDelete(t *testing.T, options globalregistry.RegistryOptions) bool {
	cfd, ok := options.(globalregistry.CanForceDelete)
	if !ok {
		t.Fatal("registryoptions is not CanForceDelete")
		return false
	}
	return cfd.ForceDeleteProjects()

}

func apiProviderName(t *testing.T, ap *mockApiProvider) string {
	switch ap {
	default:
		t.Fatal("cannot get the name of mockApiProvider")
		return ""
	case apiProviderForceDelete:
		return "apiProvider: forceDelete"
	case apiProviderNoForceDelete:
		return "apiProvider: noForceDelete"
	}
}

func apiRegistryName(t *testing.T, reg *api.Registry) string {
	switch reg {
	default:
		t.Fatal("cannot get the name of api.Registry")
		return ""
	case registryForceDelete:
		return "registry: forceDelete"
	case registryNoForceDelete:
		return "registry: noForceDelete"
	case registryInvalidForceDelete:
		return "registry: invalidForceDelete"
	case registryMissingForceDelete:
		return "registry: missingForceDelete"
	}
}

func testGetOptions(t *testing.T, ap *mockApiProvider, apiReg *api.Registry, optionName string) (bool, error) {
	reg := New(apiReg, ap)
	t.Logf("%s, %s",
		apiProviderName(t, ap),
		apiRegistryName(t, apiReg),
	)
	switch optionName {
	case "forceDelete":
		return canForceDelete(t, reg.GetOptions()), nil
	default:
		return false, fmt.Errorf("invalid optionName: %v", optionName)
	}
}

func TestRegistry_GetOptions(t *testing.T) {

	registryTest := []struct {
		id         string
		cliOption  *mockApiProvider
		apiOption  *api.Registry
		optionName string
		expResult  bool
	}{
		{id: "1", cliOption: apiProviderForceDelete, apiOption: registryForceDelete, optionName: "forceDelete", expResult: true},
		{id: "2", cliOption: apiProviderForceDelete, apiOption: registryNoForceDelete, optionName: "forceDelete", expResult: false},
		{id: "3", cliOption: apiProviderForceDelete, apiOption: registryInvalidForceDelete, optionName: "forceDelete", expResult: false},
		{id: "4", cliOption: apiProviderForceDelete, apiOption: registryMissingForceDelete, optionName: "forceDelete", expResult: true},
		{id: "5", cliOption: apiProviderNoForceDelete, apiOption: registryForceDelete, optionName: "forceDelete", expResult: true},
		{id: "6", cliOption: apiProviderNoForceDelete, apiOption: registryNoForceDelete, optionName: "forceDelete", expResult: false},
		{id: "7", cliOption: apiProviderNoForceDelete, apiOption: registryInvalidForceDelete, optionName: "forceDelete", expResult: false},
		{id: "8", cliOption: apiProviderNoForceDelete, apiOption: registryMissingForceDelete, optionName: "forceDelete", expResult: false},
	}

	for _, tt := range registryTest {
		t.Run(tt.id, func(t *testing.T) {
			got, err := testGetOptions(t, tt.cliOption, tt.apiOption, tt.optionName)
			if err != nil {
				t.Fatalf("%v", err)
			}
			if got != tt.expResult {
				t.Errorf("TC-%v got %t want %t", tt.id, got, tt.expResult)
			}
		})
	}
}

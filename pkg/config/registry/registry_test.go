package registry

import (
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman.kubermatic.com/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockApiProvider struct {
	forceDelete bool
}

func (ap *mockApiProvider) GetProjects() []*api.Project                   { return nil }
func (ap *mockApiProvider) GetRegistries() []*api.Registry                { return nil }
func (ap *mockApiProvider) GetScanners() []*api.Scanner                   { return nil }
func (ap *mockApiProvider) GetCliOptions() globalregistry.RegistryOptions { return ap }
func (ap *mockApiProvider) GetLogger() logr.Logger                        { return logger }
func (ap *mockApiProvider) ForceDeleteProjects() bool                     { return ap.forceDelete }

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

func testGetOptions(t *testing.T, ap *mockApiProvider, apiReg *api.Registry, expResult bool) {
	reg := New(apiReg, ap)
	t.Logf("%s, %s",
		apiProviderName(t, ap),
		apiRegistryName(t, apiReg),
	)
	if canForceDelete(t, reg.GetOptions()) != expResult {
		t.Error("unexpected result")
	}
}
func TestRegistry_GetOptions(t *testing.T) {
	testGetOptions(t, apiProviderForceDelete, registryForceDelete, true)
	testGetOptions(t, apiProviderForceDelete, registryNoForceDelete, false)
	testGetOptions(t, apiProviderForceDelete, registryInvalidForceDelete, false)
	testGetOptions(t, apiProviderForceDelete, registryMissingForceDelete, true)

	testGetOptions(t, apiProviderNoForceDelete, registryForceDelete, true)
	testGetOptions(t, apiProviderNoForceDelete, registryNoForceDelete, false)
	testGetOptions(t, apiProviderNoForceDelete, registryInvalidForceDelete, false)
	testGetOptions(t, apiProviderNoForceDelete, registryMissingForceDelete, false)
}

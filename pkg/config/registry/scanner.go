package registry

import (
	api "github.com/kubermatic-labs/registryman/pkg/apis/globalregistry/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type scanner struct {
	*api.Scanner
}

var _ globalregistry.Scanner = &scanner{}

func (s *scanner) GetURL() string {
	return s.Spec.Url
}

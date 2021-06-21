package harbor

import (
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type scanner struct {
	id        string
	api       *scannerAPI
	name      string
	url       string
	isDefault bool
}

var _ globalregistry.Scanner = &scanner{}

func (s *scanner) Delete() error {
	return s.api.delete(s.id)
}

func (s *scanner) GetName() string {
	return s.name
}

func (s *scanner) GetURL() string {
	return s.url
}

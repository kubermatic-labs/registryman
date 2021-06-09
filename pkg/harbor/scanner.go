package harbor

import (
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type scanner struct {
	id   string
	api  *scannerAPI
	Name string
}

var _ globalregistry.Scanner = &scanner{}

func (s *scanner) Delete() error {
	return s.api.delete(s.id)
}

func (s *scanner) GetRegistrationID() string {
	return s.id
}

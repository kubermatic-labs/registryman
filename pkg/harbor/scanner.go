package harbor

import (
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

type Scanner struct {
	id   string
	api  *scannerAPI
	name string
	url  string
}

var _ globalregistry.Scanner = &Scanner{}

func (s *Scanner) Delete() error {
	return s.api.delete(s.id)
}

func (s *Scanner) GetName() string {
	return s.name
}

func (s *Scanner) GetURL() string {
	return s.url
}

func (s *Scanner) getID() string {
	return s.id
}

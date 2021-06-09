package harbor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
)

const scannersPath = "/api/v2.0/scanners"

type scannerRegistration struct {
	Disabled         bool     `json:"disabled,omitempty"`
	Vendor           string   `json:"vendor,omitempty"`
	Description      string   `json:"description,omitempty"`
	Url              *url.URL `json:"url,omitempty"`
	Adapter          string   `json:"adapter,omitempty"`
	AccessCredential string   `json:"access_credential,omitempty"`
	Uuid             string   `json:"uuid,omitempty"`
	Auth             string   `json:"auth,omitempty"`
	IsDefault        bool     `json:"is_default,omitempty"`
	Version          string   `json:"version,omitempty"`
	Health           string   `json:"health,omitempty"`
	UseInternalAddr  bool     `json:"use_internal_addr,omitempty"`
	SkipCertVerify   bool     `json:"skip_cert_verify,omitempty"`
	Name             string   `json:"name,omitempty"`
}

type scannerAPI struct {
	reg *registry
}

var _ globalregistry.ScannerAPI = &scannerAPI{}

func newScannerAPI(reg *registry) (*scannerAPI, error) {
	return &scannerAPI{
		reg: reg,
	}, nil
}

func (s *scannerAPI) Create(name string) (globalregistry.Scanner, error) {
	return nil, fmt.Errorf("scannerAPI.Create() is not implemented")
}
func (s *scannerAPI) SetDefaultSystemScanner(globalregistry.Scanner) error {
	return fmt.Errorf("scannerAPI.SetDefaultSystemScanner is not implemented")
}
func (s *scannerAPI) List() ([]globalregistry.Scanner, error) {
	s.reg.parsedUrl.Path = scannersPath
	req, err := http.NewRequest(http.MethodGet, s.reg.parsedUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(s.reg.GetUsername(), s.reg.GetPassword())
	resp, err := s.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	scannerResult := []*scannerRegistration{}
	err = json.NewDecoder(resp.Body).Decode(&scannerResult)

	if err != nil {
		s.reg.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		s.reg.logger.Info(b.String())
	}

	scanners := make([]globalregistry.Scanner, 0)

	if len(scannerResult) == 0 {
		return scanners, err
	}

	for _, scannerIterator := range scannerResult {
		scanners = append(scanners, &scanner{
			id:   scannerIterator.Uuid,
			api:  s,
			Name: scannerIterator.Name,
		})
	}
	return scanners, err
}

func (s *scannerAPI) GetForProject(id int) (globalregistry.Scanner, error) {
	s.reg.parsedUrl.Path = fmt.Sprintf("%s/%d/scanner", path, id)
	req, err := http.NewRequest(http.MethodGet, s.reg.parsedUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(s.reg.GetUsername(), s.reg.GetPassword())
	resp, err := s.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	scannerResult := &scannerRegistration{}
	err = json.NewDecoder(resp.Body).Decode(scannerResult)

	if err != nil {
		s.reg.logger.Error(err, "json decoding failed")
		b := bytes.NewBuffer(nil)
		_, err := b.ReadFrom(resp.Body)
		if err != nil {
			panic(err)
		}
		s.reg.logger.Info(b.String())
	}

	scanner := &scanner{
		id:   scannerResult.Uuid,
		api:  s,
		Name: scannerResult.Name,
	}
	return scanner, err

}

func (s *scannerAPI) delete(id string) error {
	return fmt.Errorf("scannerAPI.delete() is not implemented")
}

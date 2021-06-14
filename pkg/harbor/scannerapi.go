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
	Disabled         bool   `json:"disabled,omitempty"`
	Vendor           string `json:"vendor,omitempty"`
	Description      string `json:"description,omitempty"`
	Url              string `json:"url,omitempty"`
	Adapter          string `json:"adapter,omitempty"`
	AccessCredential string `json:"access_credential,omitempty"`
	Uuid             string `json:"uuid,omitempty"`
	Auth             string `json:"auth,omitempty"`
	IsDefault        bool   `json:"is_default,omitempty"`
	Version          string `json:"version,omitempty"`
	Health           string `json:"health,omitempty"`
	UseInternalAddr  bool   `json:"use_internal_addr,omitempty"`
	SkipCertVerify   bool   `json:"skip_cert_verify,omitempty"`
	Name             string `json:"name,omitempty"`
}

type scannerRegistrationRequest struct {
	Name             string `json:"name,omitempty"`
	Url              string `json:"url,omitempty"`
	AccessCredential string `json:"access_credential,omitempty"`
	Auth             string `json:"auth,omitempty"`
	Disabled         bool   `json:"disabled,omitempty"`
	UseInternalAddr  bool   `json:"use_internal_addr,omitempty"`
	SkipCertVerify   bool   `json:"skip_cert_verify,omitempty"`
	Description      string `json:"description,omitempty"`
}

type projectScanner struct {
	Uuid string `json:"uuid,omitempty"`
}
type scannerAPI struct {
	reg *registry
}

var _ globalregistry.ScannerConfig = &scannerRegistrationRequest{}

func newScannerAPI(reg *registry) *scannerAPI {
	return &scannerAPI{
		reg: reg,
	}
}

func (s *scannerAPI) Create(config globalregistry.ScannerConfig) (*url.URL, error) {
	s.reg.parsedUrl.Path = scannersPath

	reqBodyBuf := bytes.NewBuffer(nil)
	err := json.NewEncoder(reqBodyBuf).Encode(&scannerRegistrationRequest{
		Name:             config.GetName(),
		Url:              config.GetUrl(),
		AccessCredential: config.GetCredential(),
		Auth:             config.GetAuth(),
		Description:      config.GetDescription(),
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, s.reg.parsedUrl.String(), reqBodyBuf)
	if err != nil {
		return nil, err
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(s.reg.GetUsername(), s.reg.GetPassword())
	resp, err := s.reg.do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("scanner creation failed, %w", globalregistry.RecoverableError)
	}

	scannerUrl := resp.Header.Get("Location")
	parsedUrl, err := url.Parse(scannerUrl)

	if err != nil {
		s.reg.logger.Error(err, "cannot parse scanner URL from response Location header",
			"location-header", resp.Header.Get("Location"))
		return nil, err
	}
	return parsedUrl, nil
}

func (s *scannerAPI) getScannerIDByName(name string) (string, error) {
	scanners, err := s.List()
	if err != nil {
		return "", err
	}
	for _, scanner := range scanners {
		if scanner.GetName() == name {
			return scanner.(*Scanner).getID(), err
		}
	}
	return "", nil
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
		scanners = append(scanners, &Scanner{
			id:   scannerIterator.Uuid,
			api:  s,
			name: scannerIterator.Name,
		})
	}
	return scanners, err
}

func (s *scannerAPI) SetForProject(projectID int, scannerID string) error {
	s.reg.parsedUrl.Path = fmt.Sprintf("%s/%d/scanner", path, projectID)

	reqBodyBuf := bytes.NewBuffer(nil)
	err := json.NewEncoder(reqBodyBuf).Encode(&projectScanner{
		Uuid: scannerID,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, s.reg.parsedUrl.String(), reqBodyBuf)
	if err != nil {
		return err
	}

	req.SetBasicAuth(s.reg.GetUsername(), s.reg.GetPassword())
	resp, err := s.reg.do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to set scanner for project-id:%d, %w", projectID, globalregistry.RecoverableError)
	}
	return err
}

func (s *scannerAPI) getForProject(id int) (globalregistry.Scanner, error) {
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

	scanner := &Scanner{
		id:   scannerResult.Uuid,
		api:  s,
		name: scannerResult.Name,
		url:  scannerResult.Url,
	}
	return scanner, err

}

func (s *scannerAPI) delete(id string) error {
	s.reg.parsedUrl.Path = fmt.Sprintf("%s/%s", scannersPath, id)

	req, err := http.NewRequest(http.MethodDelete, s.reg.parsedUrl.String(), nil)
	if err != nil {
		return err
	}

	req.Header["Content-Type"] = []string{"application/json"}
	req.SetBasicAuth(s.reg.GetUsername(), s.reg.GetPassword())

	resp, err := s.reg.do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to remove scanner, %w", globalregistry.RecoverableError)
	}
	return err
}

func (c *scannerRegistrationRequest) GetName() string {
	return c.Name
}

func (c *scannerRegistrationRequest) GetAuth() string {
	return c.Auth
}

func (c *scannerRegistrationRequest) GetCredential() string {
	return c.AccessCredential
}

func (c *scannerRegistrationRequest) GetUrl() string {
	return c.Url
}

func (c *scannerRegistrationRequest) IsDisabled() bool {
	return c.Disabled
}

func (c *scannerRegistrationRequest) GetDescription() string {
	return c.Description
}

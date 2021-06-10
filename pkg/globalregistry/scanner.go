package globalregistry

import "net/url"

type Scanner interface {
	GetID() string
	Delete() error
}

type ScannerAPI interface {
	Create(config ScannerConfig) (*url.URL, error)
	// TODO: Is it needed?
	SetForProject(projectID int, scannerID string) error
	GetForProject(projectID int) (Scanner, error)
	List() ([]Scanner, error)
}

type ScannerConfig interface {
	GetName() string
	GetUrl() url.URL
	GetCredential() string
	GetAuth() string
	IsDisabled() bool
	GetDescription() string
}

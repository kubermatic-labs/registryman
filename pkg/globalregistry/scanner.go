package globalregistry

type Scanner interface {
	GetName() string
	GetURL() string
	// GetID() string
	// Delete() error
}

// type ScannerAPI interface {
// 	Create(config ScannerConfig) (*url.URL, error)
// 	// TODO: Is it needed?
// 	SetForProject(projectID int, scannerID string) error
// 	GetForProject(projectID int) (Scanner, error)
// 	List() ([]Scanner, error)
// }

type ScannerConfig interface {
	GetName() string
	GetUrl() string
	GetCredential() string
	GetAuth() string
	IsDisabled() bool
	GetDescription() string
}

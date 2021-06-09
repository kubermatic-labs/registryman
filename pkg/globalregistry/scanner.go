package globalregistry

type Scanner interface {
	GetRegistrationID() string
	Delete() error
}

type ScannerAPI interface {
	Create(name string) (Scanner, error)
	SetDefaultSystemScanner(Scanner) error
	GetForProject(id int) (Scanner, error)
	List() ([]Scanner, error)
}

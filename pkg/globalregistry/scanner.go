package globalregistry

type Scanner interface {
	GetName() string
	GetURL() string
}

type ScannerConfig interface {
	GetName() string
	GetUrl() string
}

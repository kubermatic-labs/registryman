package globalregistry

// Storage interface contains the methods that we use for storage related
// operations. If the provider does not implement the GetUsedStorage, it shall
// return -1, ErrNotImplemented
type Storage interface {
	// GetUsedStorage returns the used storage in bytes.
	GetUsedStorage() (int, error)
}

package globalregistry

// CanForceDelete interface describes an option that is needed to
// be able to delete a project when it has repositories in it.
type CanForceDelete interface {
	ForceDeleteProjects() bool
}

// RegistryOptions interface describes the registry options
// coming from CLI options, or from the registry description.
type RegistryOptions interface {
}

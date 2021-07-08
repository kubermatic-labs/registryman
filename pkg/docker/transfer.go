package docker

import (
	"fmt"

	"github.com/containers/image/v5/directory"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
	"github.com/go-logr/logr"
)

type transfer struct {
	dockerCtx *types.SystemContext
	dirCtx    *types.SystemContext
}
type transferData struct {
	sourcePath           string
	destinationPath      string
	sourceCtx            *types.SystemContext
	destinationCtx       *types.SystemContext
	sourceTransport      string
	destinationTransport string
	scoped               bool
}

func New(username, password string) *transfer {
	return &transfer{
		dockerCtx: &types.SystemContext{
			DockerAuthConfig: &types.DockerAuthConfig{
				Username: username,
				Password: password,
			},
		},
		dirCtx: &types.SystemContext{},
	}
}

// os-images -> "harbor-1.lab.kubermatic.io/os-images/photon"
func (t *transfer) Export(source, destination string, logger logr.Logger) error {
	logger.Info("exporting images started")

	err := syncImages(&transferData{
		sourcePath:           source,
		destinationPath:      destination,
		sourceCtx:            t.dockerCtx,
		destinationCtx:       t.dirCtx,
		sourceTransport:      docker.Transport.Name(),
		destinationTransport: directory.Transport.Name(),
		scoped:               true,
	})

	if err != nil {
		return fmt.Errorf("syncing images failed: %w", err)
	}

	return nil
}

func (t *transfer) Import(source, destination string, logger logr.Logger) error {
	logger.Info("importing images started")

	err := syncImages(&transferData{
		sourcePath:           source,
		destinationPath:      destination,
		sourceCtx:            t.dirCtx,
		destinationCtx:       t.dockerCtx,
		sourceTransport:      directory.Transport.Name(),
		destinationTransport: docker.Transport.Name(),
		scoped:               false,
	})

	if err != nil {
		return fmt.Errorf("syncing images failed: %w", err)
	}

	return nil
}

func (t *transfer) Sync(sourceRepo, destinationRepo string, logger logr.Logger) error {
	logger.Info("syncing images started")
	err := syncImages(&transferData{
		sourcePath:           sourceRepo,
		destinationPath:      destinationRepo,
		sourceCtx:            t.dockerCtx,
		destinationCtx:       t.dockerCtx,
		sourceTransport:      docker.Transport.Name(),
		destinationTransport: docker.Transport.Name(),
		scoped:               false,
	})

	if err != nil {
		return fmt.Errorf("syncing images failed: %w", err)
	}

	return nil
}

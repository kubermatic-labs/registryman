package docker

import (
	"fmt"

	"github.com/containers/image/v5/directory"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
	"github.com/go-logr/logr"
)

const (
	dockerTransportPrefix    = "docker://"
	directoryTransportPrefix = "dir:"
)

var (
	dockerCtx = &types.SystemContext{
		DockerAuthConfig: &types.DockerAuthConfig{
			Username: "admin",
			Password: "P8wk%pU9D!#bSp",
		},
	}
	dirCtx = &types.SystemContext{}
)

type transferData struct {
	sourcePath           string
	destinationPath      string
	sourceCtx            *types.SystemContext
	destinationCtx       *types.SystemContext
	sourceTransport      string
	destinationTransport string
	scoped               bool
}

// os-images -> "harbor-1.lab.kubermatic.io/os-images/photon"
func Export(source, destination string, logger logr.Logger) error {
	logger.Info("exporting images started")

	// err := exportImages(
	// 	fmt.Sprintf("%s%s", dockerTransportPrefix, source),
	// 	fmt.Sprintf("%s%s", directoryTransportPrefix, destination),
	// 	sourceCtx,
	// )
	// if err != nil {
	// 	return fmt.Errorf("exporting images failed: %w", err)
	// }

	err := syncImages(&transferData{
		sourcePath:           source,
		destinationPath:      destination,
		sourceCtx:            dockerCtx,
		destinationCtx:       dirCtx,
		sourceTransport:      docker.Transport.Name(),
		destinationTransport: directory.Transport.Name(),
		scoped:               true,
	})

	if err != nil {
		return fmt.Errorf("syncing images failed: %w", err)
	}

	return nil
}

func Import(source, destination string, logger logr.Logger) error {
	logger.Info("importing images started")

	err := syncImages(&transferData{
		sourcePath:           source,
		destinationPath:      destination,
		sourceCtx:            dirCtx,
		destinationCtx:       dockerCtx,
		sourceTransport:      directory.Transport.Name(),
		destinationTransport: docker.Transport.Name(),
		scoped:               false,
	})

	if err != nil {
		return fmt.Errorf("syncing images failed: %w", err)
	}

	return nil
}

func Sync(sourceRepo, destinationRepo string, logger logr.Logger) error {
	logger.Info("syncing images started")
	err := syncImages(&transferData{
		sourcePath:           sourceRepo,
		destinationPath:      destinationRepo,
		sourceCtx:            dockerCtx,
		destinationCtx:       dockerCtx,
		sourceTransport:      docker.Transport.Name(),
		destinationTransport: docker.Transport.Name(),
		scoped:               false,
	})

	if err != nil {
		return fmt.Errorf("syncing images failed: %w", err)
	}

	return nil
}

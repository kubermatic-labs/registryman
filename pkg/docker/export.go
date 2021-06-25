package docker

import (
	"fmt"

	"github.com/containers/image/v5/types"
	"github.com/go-logr/logr"
)

const (
	dockerTransportPrefix    = "docker://"
	directoryTransportPrefix = "dir:"
)

func Export(source, destination string, logger logr.Logger) error {
	logger.Info("exporting images started")

	sourceCtx := &types.SystemContext{
		DockerAuthConfig: &types.DockerAuthConfig{
			Username: "admin",
			Password: "P8wk%pU9D!#bSp",
		},
	}

	err := exportImages(
		fmt.Sprintf("%s%s", dockerTransportPrefix, source),
		fmt.Sprintf("%s%s", directoryTransportPrefix, destination),
		sourceCtx,
	)
	if err != nil {
		return fmt.Errorf("exporting images failed: %w", err)
	}

	// logger.Info("deleting images started")
	// err = deleteImage(source, sourceCtx)
	// if err != nil {
	// 	return fmt.Errorf("failed deleting image: %w", err)
	// }

	return nil
}

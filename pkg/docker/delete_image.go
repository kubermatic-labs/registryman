package docker

import (
	"context"

	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
)

func deleteImage(source string, sourceCtx *types.SystemContext) error {
	sourceRef, err := alltransports.ParseImageName(source)
	if err != nil {
		return err
	}

	ctx := context.Background()

	if err != nil {
		return err
	}

	return sourceRef.DeleteImage(ctx, sourceCtx)
}

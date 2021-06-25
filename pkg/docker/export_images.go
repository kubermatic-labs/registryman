package docker

import (
	"context"
	"fmt"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
)

func exportImages(source, destination string, sourceCtx *types.SystemContext) error {
	sourceRef, err := alltransports.ParseImageName(source)
	if err != nil {
		return err
	}

	destRef, err := alltransports.ParseImageName(destination)
	if err != nil {
		return err
	}

	ctx := context.Background()

	policy := &signature.Policy{
		Default: []signature.PolicyRequirement{
			signature.NewPRInsecureAcceptAnything(),
		},
	}
	policyContext, err := signature.NewPolicyContext(policy)
	if err != nil {
		return err
	}

	manifests, err := copy.Image(
		ctx, policyContext, destRef, sourceRef,
		&copy.Options{SourceCtx: sourceCtx},
	)
	if err != nil {
		return err
	}

	fmt.Printf("%s", string(manifests))
	return nil
}

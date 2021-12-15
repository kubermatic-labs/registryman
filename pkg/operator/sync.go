package operator

import (
	"context"
	"errors"
	"github.com/kubermatic-labs/registryman/pkg/config"
	"github.com/kubermatic-labs/registryman/pkg/config/registry"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry/reconciler"
)

type SyncableResources interface {
	registry.ApiObjectProvider
	reconciler.SideEffectPerformer
}

func resyncRegistryByName(ctx context.Context, sres SyncableResources, registryName string) error {
	expectedProvider := config.NewExpectedProvider(sres)
	expectedRegistry := expectedProvider.GetRegistryByName(ctx, registryName)
	return resyncRegistry(ctx, sres, expectedProvider, expectedRegistry, false)
}

func resyncRegistry(ctx context.Context, sres SyncableResources, expectedProvider *config.ExpectedProvider, expectedRegistry *registry.Registry, dryRun bool) error {
	logger.Info("inspecting registry", "registry_name", expectedRegistry.GetName())
	regStatusExpected, err := reconciler.GetRegistryStatus(ctx, expectedRegistry)
	if err != nil {
		return err
	}
	logger.V(1).Info("expected registry status acquired", "status", regStatusExpected)
	actualRegistry, err := expectedRegistry.ToReal()
	if err != nil {
		return err
	}
	regStatusActual, err := reconciler.GetRegistryStatus(ctx, actualRegistry)
	if err != nil {
		return err
	}
	logger.V(1).Info("actual registry status acquired", "status", regStatusActual)
	actions := reconciler.Compare(expectedProvider, regStatusActual, regStatusExpected)
	logger.Info("ACTIONS:")
	for _, action := range actions {
		if !dryRun {
			logger.Info(action.String())
			sideEffect, err := action.Perform(ctx, actualRegistry)
			if err != nil {
				if errors.Is(err, globalregistry.ErrRecoverableError) {
					logger.V(-1).Info(err.Error())
				} else {
					return err
				}
			}
			if err = sideEffect.Perform(ctx, sres); err != nil {
				return err
			}
		} else {
			logger.Info(action.String(), "dry-run", dryRun)
		}
	}

	return nil
}

// FullResync performs a complete state synchronization over all provisioned
// Registry resources.
func FullResync(ctx context.Context, aop SyncableResources, dryRun bool) error {
	expectedProvider := config.NewExpectedProvider(aop)
	for _, expectedRegistry := range expectedProvider.GetRegistries(ctx) {
		err := resyncRegistry(ctx, aop, expectedProvider, expectedRegistry, dryRun)
		if err != nil {
			return err
		}
	}

	return nil
}

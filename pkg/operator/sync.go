/*
   Copyright 2021 The Kubermatic Kubernetes Platform contributors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

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

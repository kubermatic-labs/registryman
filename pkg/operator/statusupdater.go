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
	"time"

	api "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1"
	"github.com/kubermatic-labs/registryman/pkg/config/registry"
	"github.com/kubermatic-labs/registryman/pkg/globalregistry/reconciler"
)

type RegistryStore interface {
	registry.ApiObjectProvider

	// UpdateRegistryStatus persists the registry status of the given
	// Registry resource.
	UpdateRegistryStatus(context.Context, *api.Registry) error
}

type StatusUpdater struct {
	interval time.Duration
	store    RegistryStore
}

func NewStatusUpdater(interval time.Duration, store RegistryStore) *StatusUpdater {
	return &StatusUpdater{
		interval: interval,
		store:    store,
	}
}

func (sup *StatusUpdater) Start(ctx context.Context) {
	logger.V(1).Info("starting statusupdater")
	go sup.loop(ctx)
}

func (sup *StatusUpdater) loop(ctx context.Context) {
	timer := time.NewTicker(sup.interval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.V(1).Info("stopping statusupdater loop")
			return
		case <-timer.C:
			logger.V(1).Info("statusupdater tick")
			registries := sup.store.GetRegistries(ctx)
			for _, registry := range registries {
				go sup.updateRegistryStatus(ctx, registry)
			}
		}
	}
}

func (sup *StatusUpdater) updateRegistryStatus(ctx context.Context, reg *api.Registry) {
	logger.V(1).Info("updating registry status",
		"registry", reg.GetName(),
	)
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, sup.interval)
	defer cancel()
	realReg, err := registry.New(reg, sup.store).ToReal()
	if err != nil {
		logger.Error(err, "failed to create a real registry")
		return
	}
	registryStatus, err := reconciler.GetRegistryStatus(ctx, realReg)
	logger.V(1).Info("getting registrystatus",
		"registry", reg.GetName(),
		"status", registryStatus,
	)
	if err != nil {
		logger.Error(err, "failed getting registry status in statusupdater")
		return
	}
	reg.Status = registryStatus
	err = sup.store.UpdateRegistryStatus(ctx, reg)
	if err != nil {
		logger.Error(err, "failed updating registry status in statusupdater")
		return
	}
}

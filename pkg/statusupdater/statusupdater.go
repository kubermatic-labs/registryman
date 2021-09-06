package statusupdater

import (
	"context"
	"time"

	"github.com/go-logr/logr"
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
	logger   logr.Logger
	interval time.Duration
	store    RegistryStore
}

func New(logger logr.Logger, interval time.Duration, store RegistryStore) *StatusUpdater {
	return &StatusUpdater{
		logger:   logger,
		interval: interval,
		store:    store,
	}
}

func (sup *StatusUpdater) Start(ctx context.Context) {
	sup.logger.V(1).Info("starting statusupdater")
	go sup.loop(ctx)
}

func (sup *StatusUpdater) loop(ctx context.Context) {
	timer := time.NewTicker(sup.interval)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			sup.logger.V(1).Info("stopping statusupdater loop")
			return
		case <-timer.C:
			sup.logger.V(1).Info("statusupdater tick")
			registries := sup.store.GetRegistries(ctx)
			for _, registry := range registries {
				go sup.updateRegistryStatus(ctx, registry)
			}
		}
	}
}

func (sup *StatusUpdater) updateRegistryStatus(ctx context.Context, reg *api.Registry) {
	sup.logger.V(1).Info("updating registry status",
		"registry", reg.GetName(),
	)
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, sup.interval)
	defer cancel()
	realReg, err := registry.New(reg, sup.store).ToReal()
	if err != nil {
		sup.logger.Error(err, "failed to create a real registry")
		return
	}
	registryStatus, err := reconciler.GetRegistryStatus(ctx, realReg)
	sup.logger.V(1).Info("getting registrystatus",
		"registry", reg.GetName(),
		"status", registryStatus,
	)
	if err != nil {
		sup.logger.Error(err, "failed getting registry status in statusupdater")
		return
	}
	reg.Status = registryStatus
	err = sup.store.UpdateRegistryStatus(ctx, reg)
	if err != nil {
		sup.logger.Error(err, "failed updating registry status in statusupdater")
		return
	}
}

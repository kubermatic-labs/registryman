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

	regmaninformer "github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/informers/externalversions"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var defaultResync = 30 * time.Second

type EventRecorder interface {
	RecordEventNormal(obj runtime.Object, reason, message string)
	RecordEventWarning(obj runtime.Object, reason, message string)
}

type AOSWithSharedInformerFactory interface {
	SyncableResources
	EventRecorder

	// SharedInformerFactory returns a SharedInformerFactory.
	SharedInformerFactory(defaultResync time.Duration) regmaninformer.SharedInformerFactory
}

//Reconciler type is responsible for the registryman reconciliation loop.
type Reconciler struct {
	aos AOSWithSharedInformerFactory
}

func NewReconciler(aos AOSWithSharedInformerFactory) *Reconciler {
	return &Reconciler{
		aos: aos,
	}
}

func (rec *Reconciler) Start(ctx context.Context) {
	logger.V(1).Info("starting reconciler")
	go rec.loop(ctx)
}

func (rec *Reconciler) loop(ctx context.Context) {
	logger.V(1).Info("creating shared informer factory",
		"defaultResync", defaultResync)
	siFactory := rec.aos.SharedInformerFactory(defaultResync)
	registryInformer, err := siFactory.ForResource(schema.GroupVersionResource{
		Group:    "registryman.kubermatic.com",
		Version:  "v1alpha1",
		Resource: "registries",
	})
	if err != nil {
		logger.Error(err, "cannot create registryInformer")
		return
	}
	registryInformer.Informer().AddEventHandler(
		&registryEventHandler{
			ctx:    ctx,
			aop:    rec.aos,
			events: rec.aos,
		})
	go func() {
		registryInformer.Informer().Run(ctx.Done())
		logger.V(-1).Info("registry informer stopped")
		panic("registry informer stopped")
	}()

	projectInformer, err := siFactory.ForResource(schema.GroupVersionResource{
		Group:    "registryman.kubermatic.com",
		Version:  "v1alpha1",
		Resource: "projects",
	})
	if err != nil {
		logger.Error(err, "cannot create projectInformer")
		return
	}
	projectInformer.Informer().AddEventHandler(
		&projectEventHandler{
			ctx:    ctx,
			aop:    rec.aos,
			events: rec.aos,
		})
	go func() {
		projectInformer.Informer().Run(ctx.Done())
		logger.V(-1).Info("project informer stopped")
		panic("project informer stopped")
	}()

	scannerInformer, err := siFactory.ForResource(schema.GroupVersionResource{
		Group:    "registryman.kubermatic.com",
		Version:  "v1alpha1",
		Resource: "scanners",
	})
	if err != nil {
		logger.Error(err, "cannot create scannerInformer")
		return
	}
	scannerInformer.Informer().AddEventHandler(
		&scannerEventHandler{
			ctx:    ctx,
			aop:    rec.aos,
			events: rec.aos,
		})
	go func() {
		scannerInformer.Informer().Run(ctx.Done())
		logger.V(-1).Info("scanner informer stopped")
		panic("scanner informer stopped")
	}()
	<-ctx.Done()
	logger.V(1).Info("stopping reconciler loop")
}

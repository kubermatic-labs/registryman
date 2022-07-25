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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/cache"
)

type registryEventHandler struct {
	ctx    context.Context
	aop    SyncableResources
	events EventRecorder
}

var _ cache.ResourceEventHandler = &registryEventHandler{}

func (reh *registryEventHandler) OnAdd(obj interface{}) {
	logger.V(1).Info("registryEventHander.OnAdd")
	err := FullResync(reh.ctx, reh.aop, false)
	if err != nil {
		logger.Error(err, "failed to synchronize states",
			"kind", "Registry",
			"event", "OnAdd",
		)
		reh.events.RecordEventWarning(obj.(runtime.Object),
			"RegistryUpdateFailed",
			fmt.Sprintf("Failed to synchronize states: %s", err.Error()),
		)
	}
}

func (reh *registryEventHandler) OnUpdate(oldObj, newObj interface{}) {
	logger.V(1).Info("registryEventHander.OnUpdate")
	err := FullResync(reh.ctx, reh.aop, false)
	if err != nil {
		logger.Error(err, "failed to synchronize states",
			"kind", "Registry",
			"event", "OnUpdate",
		)
		reh.events.RecordEventWarning(oldObj.(runtime.Object),
			"RegistryUpdateFailed",
			fmt.Sprintf("Failed to synchronize states: %s", err.Error()),
		)
	}
}

func (reh *registryEventHandler) OnDelete(obj interface{}) {
	logger.V(1).Info("registryEventHander.OnDelete")
	err := FullResync(reh.ctx, reh.aop, false)
	if err != nil {
		logger.Error(err, "failed to synchronize states",
			"kind", "Registry",
			"event", "OnDelete",
		)
		reh.events.RecordEventWarning(obj.(runtime.Object),
			"RegistryUpdateFailed",
			fmt.Sprintf("Failed to synchronize states: %s", err.Error()),
		)
	}
}

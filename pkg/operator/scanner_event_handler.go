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

	"k8s.io/client-go/tools/cache"
)

type scannerEventHandler struct {
	ctx context.Context
	aop SyncableResources
}

var _ cache.ResourceEventHandler = &scannerEventHandler{}

func (reh *scannerEventHandler) OnAdd(obj interface{}) {
	logger.V(1).Info("scannerEventHander.OnAdd")
	err := FullResync(reh.ctx, reh.aop, false)
	if err != nil {
		logger.Error(err, "failed to synchronize states",
			"kind", "Scanner",
			"event", "OnAdd",
		)
	}
}

func (reh *scannerEventHandler) OnUpdate(oldObj, newObj interface{}) {
	logger.V(1).Info("scannerEventHander.OnUpdate")
	err := FullResync(reh.ctx, reh.aop, false)
	if err != nil {
		logger.Error(err, "failed to synchronize states",
			"kind", "Scanner",
			"event", "OnUpdate",
		)
	}
}

func (reh *scannerEventHandler) OnDelete(obj interface{}) {
	logger.V(1).Info("scannerEventHander.OnDelete")
	err := FullResync(reh.ctx, reh.aop, false)
	if err != nil {
		logger.Error(err, "failed to synchronize states",
			"kind", "Scanner",
			"event", "OnDelete",
		)
	}
}

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

package reconciler

import (
	"context"

	"github.com/kubermatic-labs/registryman/pkg/globalregistry"
	"k8s.io/apimachinery/pkg/runtime"
)

type nilSideEffect struct{}

func (se nilSideEffect) Perform(context.Context, SideEffectPerformer) error {
	return nil
}

var nilEffect SideEffect = nilSideEffect{}

// SideEffectPerformer interface declares the methods that a SideEffect wants to
// use.
type SideEffectPerformer interface {
	WriteResource(ctx context.Context, obj runtime.Object) error
	RemoveResource(ctx context.Context, obj runtime.Object) error
}

// SideEffect interface contains the methods that a sideeffect needs to
// implement. SideEffects are optional operations that are performed after
// Actions. SideEffect can be used for e.g. file manipulations at the local
// filesystem.
type SideEffect interface {
	Perform(context.Context, SideEffectPerformer) error
}

// Action interface contains the methods that a reconciliation action needs to
// implement.
type Action interface {
	String() string
	Perform(context.Context, globalregistry.Registry) (SideEffect, error)
}

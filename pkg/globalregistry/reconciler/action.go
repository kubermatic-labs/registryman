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
)

type nilSideEffect struct{}

func (se nilSideEffect) Perform(context.Context) error {
	return nil
}

var nilEffect = nilSideEffect{}

type SideEffectContextKey string

// SideEffectManifestManipulator is a context key that may be parsed by a
// SideEffect. The value can be used by the SideEffect to create or to remove
// manifest files.
//
// The value shall implement the following interface:
//   type manifestManipulator interface {
//	WriteManifest(filename string, obj runtime.Object) error
//	RemoveManifest(filename string) error
//   }
var SideEffectManifestManipulator = SideEffectContextKey("sideeffect-manipulator")

// SideEffect interface contains the methods that a sideeffect needs to
// implement. SideEffects are optional operations that are performed after
// Actions. SideEffect can be used for e.g. file manipulations at the local
// filesystem.
type SideEffect interface {
	Perform(ctx context.Context) error
}

// Action interface contains the methods that a reconciliation action needs to
// implement.
type Action interface {
	String() string
	Perform(globalregistry.Registry) (SideEffect, error)
}

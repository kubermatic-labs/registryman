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

package globalregistry

import (
	"errors"
)

var (
	// ErrRecoverableError is an error value that indicates that the error
	// shall be logged to the user but the operation can continue.
	ErrRecoverableError error = errors.New("recoverable error")

	// ErrNotImplemented is an error value that indicates that a method is
	// not implemented by a registry provider.
	ErrNotImplemented error = errors.New("not implemented")

	// ErrAlreadyExists is an error value that indicates that the resource
	// to be created exists already.
	ErrAlreadyExists error = errors.New("already exists")
)

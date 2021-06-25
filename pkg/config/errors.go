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

package config

import "errors"

// ErrValidationInvalidLocalRegistryInProject error indicates that a local
// project refers to a non-existing registry.
var ErrValidationInvalidLocalRegistryInProject error = errors.New("validation error: project contains invalid registry name")

// ErrValidationMultipleGlobalRegistries error indicates that there are multiple
// global registries configured.
var ErrValidationMultipleGlobalRegistries error = errors.New("validation error: multiple global registries found")

// ErrValidationScannerNameNotUnique error indicates that there are multiple
// scanners configured with the same name.
var ErrValidationScannerNameNotUnique error = errors.New("validation error: multiple scanners present with the same name")

// ErrValidationProjectNameNotUnique error indicates that there are multiple
// projects configured with the same name.
var ErrValidationProjectNameNotUnique error = errors.New("validation error: multiple projects present with the same name")

// ErrValidationRegistryNameNotUnique error indicates that there are multiple
// registries configured with the same name.
var ErrValidationRegistryNameNotUnique error = errors.New("validation error: multiple registries present with the same name")

// ErrValidationScannerNameReference error indicates that a project refers to a
// non-existing Scanner.
var ErrValidationScannerNameReference error = errors.New("validation error: multiple registries present with the same name")

// ErrValidationScannerNameReference error indicates that a project refers to a
// non-existing Scanner.
var ErrValidationGroupWithoutDN error = errors.New("validation error: project group member with missing DN field")

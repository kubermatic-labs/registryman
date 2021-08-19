#!/usr/bin/env bash

# Copyright 2021 The Kubermatic Kubernetes Platform contributors.

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at

# http://www.apache.org/licenses/LICENSE-2.0

# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..

echo "Generating CRDs"
go run \
   "${SCRIPT_ROOT}"/vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go crd \
   paths="$(realpath "${SCRIPT_ROOT}"/pkg/apis/registryman/v1alpha1)"               \
   output:dir="${SCRIPT_ROOT}"/pkg/apis/registryman/v1alpha1

echo "Generating deepcopy code"
go run \
   "${SCRIPT_ROOT}/vendor/k8s.io/code-generator/cmd/deepcopy-gen/main.go"     \
   --input-dirs github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1 \
   -O zz_generated.deepcopy                                                \
   --bounding-dirs github.com/kubermatic-labs/registryman/pkg/apis         \
   -h "${SCRIPT_ROOT}/hack/boilerplate.go.txt"
# -p github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1 \

echo "Generating clientset"
go run \
   "${SCRIPT_ROOT}"/vendor/k8s.io/code-generator/cmd/client-gen/main.go     \
   --input-base '' \
   --input github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1 \
   --output-package github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/clientset \
   --clientset-name versioned \
   -h "${SCRIPT_ROOT}"/hack/boilerplate.go.txt

echo "Generating listers"
go run \
   "${SCRIPT_ROOT}"/vendor/k8s.io/code-generator/cmd/lister-gen/main.go     \
   --input-dirs github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1 \
   --output-package github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/listers \
   -h "${SCRIPT_ROOT}"/hack/boilerplate.go.txt

echo "Generating informers"
go run \
   "${SCRIPT_ROOT}"/vendor/k8s.io/code-generator/cmd/informer-gen/main.go     \
   --input-dirs github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1 \
   --versioned-clientset-package github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/clientset/versioned \
   --listers-package github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/listers \
   --output-package github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1/informers \
   -h "${SCRIPT_ROOT}"/hack/boilerplate.go.txt

echo "Generating register code"
go run \
   "${SCRIPT_ROOT}"/vendor/k8s.io/code-generator/cmd/register-gen/main.go  \
   -i github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1 \
   -p github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1 \
   -h "${SCRIPT_ROOT}"/hack/boilerplate.go.txt

echo "Generating OpenAPI"
go run \
   "${SCRIPT_ROOT}"/vendor/k8s.io/code-generator/cmd/openapi-gen/main.go                                           \
   -i github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1,k8s.io/apimachinery/pkg/apis/meta/v1 \
   -p github.com/kubermatic-labs/registryman/pkg/apis/registryman/v1alpha1                                      \
   -h "${SCRIPT_ROOT}"/hack/boilerplate.go.txt

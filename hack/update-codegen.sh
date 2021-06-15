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
# CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}
# CONTROLLER_TOOLS_PKG=${CONTROLLER_TOOLS_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/sigs.k8s.io/controller-tools 2>/dev/null || echo ../controller-tools)}

# if  ! ( command -v controller-gen > /dev/null ); then
#     echo "controller-gen not found, installing sigs.k8s.io/controller-tools"
#     pushd "${CONTROLLER_TOOLS_PKG}"
#     go install ./cmd/controller-gen
#     popd
# fi


# CLIENTSET_NAME_VERSIONED=clientset \
# CLIENTSET_PKG_NAME=clientset \
# bash "${CODEGEN_PKG}/generate-groups.sh" crd \
#            github.com/kubermatic-labs/registryman/pkg/apis \
#   "globalregistry:v1alpha1" \
#   --output-base "$(dirname "${BASH_SOURCE[0]}")/../../.." \
#   --go-header-file "${SCRIPT_ROOT}/hack/boilerplate.go.txt"

# CLIENTSET_NAME_VERSIONED=clientset \
# CLIENTSET_PKG_NAME=clientset \
# CLIENTSET_NAME_INTERNAL=internalclientset \
# bash "${CODEGEN_PKG}/generate-internal-groups.sh" deepcopy,conversion \
#   k8s.io/apiextensions-apiserver/pkg/client k8s.io/apiextensions-apiserver/pkg/apis k8s.io/apiextensions-apiserver/pkg/apis \
#   "apiextensions:v1beta1,v1" \
#   --output-base "$(dirname "${BASH_SOURCE[0]}")/../../.." \
#   --go-header-file "${SCRIPT_ROOT}/hack/boilerplate.go.txt"

echo "Generating CRDs"
go run \
   "${SCRIPT_ROOT}"/vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go crd \
   paths="$(realpath "${SCRIPT_ROOT}"/pkg/apis/globalregistry/v1alpha1)"               \
   output:dir="${SCRIPT_ROOT}"/pkg/apis/globalregistry/v1alpha1

echo "Generating deepcopy code"
go run \
   "${SCRIPT_ROOT}/vendor/k8s.io/code-generator/cmd/deepcopy-gen/main.go"     \
   -i github.com/kubermatic-labs/registryman/pkg/apis/globalregistry/v1alpha1 \
   -p github.com/kubermatic-labs/registryman/pkg/apis/globalregistry/v1alpha1 \
   -h "${SCRIPT_ROOT}/hack/boilerplate.go.txt"

echo "Generating register code"
go run \
   "${SCRIPT_ROOT}"/vendor/k8s.io/code-generator/cmd/register-gen/main.go     \
   -i github.com/kubermatic-labs/registryman/pkg/apis/globalregistry/v1alpha1 \
   -p github.com/kubermatic-labs/registryman/pkg/apis/globalregistry/v1alpha1 \
   -h "${SCRIPT_ROOT}"/hack/boilerplate.go.txt

echo "Generating OpenAPI"
go run \
   "${SCRIPT_ROOT}"/vendor/k8s.io/code-generator/cmd/openapi-gen/main.go                                           \
   -i github.com/kubermatic-labs/registryman/pkg/apis/globalregistry/v1alpha1,k8s.io/apimachinery/pkg/apis/meta/v1 \
   -p github.com/kubermatic-labs/registryman/pkg/apis/globalregistry/v1alpha1                                      \
   -h "${SCRIPT_ROOT}"/hack/boilerplate.go.txt

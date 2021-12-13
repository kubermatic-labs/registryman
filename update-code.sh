#!/usr/bin/env bash

set -euo pipefail

if [ -d "generated-code" ]
then
    chmod a+rw generated-code
    rm generated-code
fi

nix-build -A generated-code --argstr registryman-from local -o generated-code
rm -rf \
   pkg/apis/registryman/v1alpha1/registryman.kubermatic.com*.yaml \
   pkg/apis/registryman/v1alpha1/openapi_generated.go             \
   pkg/apis/registryman/v1alpha1/zz_generated*                    \
   pkg/apis/registryman/v1alpha1/clientset                        \
   pkg/apis/registryman/v1alpha1/informers                        \
   pkg/apis/registryman/v1alpha1/listers

cp -a \
   generated-code/registryman.kubermatic.com*.yaml \
   generated-code/openapi_generated.go             \
   generated-code/zz_generated*                    \
   generated-code/clientset                        \
   generated-code/informers                        \
   generated-code/listers                          \
   pkg/apis/registryman/v1alpha1

chmod a+rw -R pkg/apis/registryman/v1alpha1
rm generated-code


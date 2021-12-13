#!/bin/bash

set -e -o pipefail

TESTAUTO_FROM=local    # local or git
REGISTRYMAN_FROM=local # local or git
echo "testauto: ${TESTAUTO_FROM} registryman: ${REGISTRYMAN_FROM}"
tag=$(nix eval --raw -f default.nix  \
                --argstr testauto-from "$TESTAUTO_FROM"         \
                --argstr registryman-from "$REGISTRYMAN_FROM"   \
                --argstr registryman-git-rev "$REGISTRYMAN_REV" \
                --argstr registryman-local-path "$REGISTRYMAN"  \
                --argstr testauto-git-rev "$TESTAUTO_REV"       \
                image-tag)
IMAGE="testauto:$tag"
echo "IMAGE=$IMAGE"

docker run -ti --rm                                 \
       --network=kind                               \
       --add-host "harbor:$(dig harbor +short)"     \
       --add-host "harbor2:$(dig harbor2 +short)"   \
       -v "$(pwd)":/test                              \
       -v /var/run/docker.sock:/var/run/docker.sock \
       "${IMAGE}" "$*"

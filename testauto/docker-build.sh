#!/bin/bash

set -exo pipefail
#
# Usage:
#
# docker-build <testauto version> <registryman version>
#
# where <testauto version> can be either local or git   and
#       <registryman version> can be either local or git
#
# Example:
#    ./docker-build git git
# build a docker image where both testauto and registryman come from git.
#
# Environment variables:
#
# REGISTRYMAN_GIT_REV sets the git revision of registryman. Must be set if
# <registryman version> is git.
# REGISTRYMAN_GIT_REF sets the git ref (branch) of registryman.
# REGISTRYMAN_GIT_URL sets the git URL of registryman. Must be set if
# <registryman version> is git.
# TESTAUTO_GIT_REV sets the git revision of testauto. Must be set if
# <testauto version> is git.
# TESTAUTO_GIT_REF sets the git ref (branch) of testauto.
# TESTAUTO_GIT_URL sets the git URL of testauto. Must be set if
# <testauto version> is git.

DOCKER_IMAGE=$(nix-build --show-trace -A docker                            \
                         --argstr testauto-from "$1"                         \
                         --argstr registryman-from "$2"                      \
                         --argstr registryman-local-path "$REGISTRYMAN"      \
                         --argstr registryman-git-rev "$REGISTRYMAN_GIT_REV" \
                         --argstr registryman-git-ref "$REGISTRYMAN_GIT_REF" \
                         --argstr registryman-git-url "$REGISTRYMAN_GIT_URL" \
                         --argstr testauto-git-rev "$TESTAUTO_GIT_REV"       \
                         --argstr testauto-git-ref "$TESTAUTO_GIT_REF"       \
                         --argstr testauto-git-url "$TESTAUTO_GIT_URL"       \
            )

docker load < "$DOCKER_IMAGE"

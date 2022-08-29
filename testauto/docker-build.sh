#!/usr/bin/env bash

set -exo pipefail
#
# Usage:
#
# docker-build <testauto version>
#
# Example:
#    ./docker-build
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

DOCKER_IMAGE=$(
	nix-build --show-trace -A docker
)

docker load <"$DOCKER_IMAGE"

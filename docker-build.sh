#!/bin/bash

# Usage:
#
# docker-build <registryman version>
#
# where  <registryman version> can be either local or git
#
# Example:
#    ./docker-build git
# build a docker image where registryman comes from git.
#
# Environment variables:
#
docker load < $(nix-build -A docker \
                          --argstr registryman-from "$1" \
                          --argstr registryman-git-rev "$REGISTRYMAN_REV")

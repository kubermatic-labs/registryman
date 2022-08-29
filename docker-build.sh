#!/usr/bin/env bash

# Usage:
#
# docker-build
#
# Example:
#    ./docker-build
# build a docker image where registryman comes from git.
#
# Environment variables:
#
DOCKER_IMAGE=$(nix-build -A dockerimage)

docker load <"$DOCKER_IMAGE"

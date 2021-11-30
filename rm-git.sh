#!/bin/bash
#
nix-shell --pure -A shell                               \
          --argstr registryman-from git                 \
          --argstr registryman-git-rev "$REGISTRYMAN_REV" \
          --command "registryman $*" default.nix

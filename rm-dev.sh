#!/bin/bash
#
nix-shell --pure -A shell \
          --argstr registryman-from local \
          --show-trace \
          --command "registryman $*" default.nix

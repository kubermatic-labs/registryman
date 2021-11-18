#!/bin/bash
#
nix-shell --pure -A shell \
          --argstr registryman-from local \
          --command "registryman $*" default.nix

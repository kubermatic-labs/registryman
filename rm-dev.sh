#!/bin/bash
#
nix-shell --pure -A shell \
          --argstr registryman-from local \
          --keep KUBECONFIG \
          --show-trace \
          --command "registryman $*" default.nix

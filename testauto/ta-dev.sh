#!/bin/bash
#
# The REGISTRYMAN environment variable specifies the absolute path of the local
# registryman repo.
#
# e.g. REGISTRYMAN=/path/to/registryman
nix-shell --pure -A shell \
          --argstr testauto-from local    \
          --argstr registryman-from local \
          --command "racket -l testauto -- $*"

#!/bin/bash
#
echo "Testauto git revision: ${TESTAUTO_GIT_REV}"
echo "Registryman git revision: ${REGISTRYMAN_GIT_REV}"
nix-shell --pure -A shell \
          --argstr testauto-from git                          \
          --argstr registryman-from git                       \
          --argstr registryman-git-rev "$REGISTRYMAN_GIT_REV" \
          --argstr registryman-git-ref "$REGISTRYMAN_GIT_REF" \
          --argstr registryman-git-url "$REGISTRYMAN_GIT_URL" \
          --argstr testauto-git-rev "$TESTAUTO_GIT_REV"       \
          --argstr testauto-git-ref "$TESTAUTO_GIT_REF"       \
          --argstr testauto-git-url "$TESTAUTO_GIT_URL"       \
          --command "racket -l testauto -- $*"


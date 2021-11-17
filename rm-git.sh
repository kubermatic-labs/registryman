#!/bin/bash
#
nix-shell --pure -A git --command "registryman $*" default.nix

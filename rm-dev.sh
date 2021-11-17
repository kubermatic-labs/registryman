#!/bin/bash
#
nix-shell --pure -A dev --command "registryman $*" default.nix 

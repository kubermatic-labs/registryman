#!/usr/bin/env bash
go build -v -tags "exclude_graphdriver_devicemapper exclude_graphdriver_btrfs containers_image_openpgp" .

#!/usr/bin/env bash

set -euo pipefail

eval `go run ./test/makekey ./var/foo.gpg`
eval `go run ./cmd/yeet --fname ./test/rpm-installtest.yeetfile.js --gpg-key-file=${GPG_KEY_FILE} --gpg-key-id=${GPG_KEY_ID} --force-git-version 1.0.0`

images='rockylinux/rockylinux:10
rockylinux/rockylinux:9
almalinux:10
almalinux:9
oraclelinux:10
oraclelinux:9
fedora:41
fedora:42
fedora:43'

for baseImage in $images; do
  echo $baseImage
  docker build \
    --build-arg IMAGE=$baseImage \
    --build-arg PACKAGE=${PACKAGE_PATH} \
    --file test/rpm-installtest.Dockerfile \
    .
done
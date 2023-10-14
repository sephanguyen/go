#!/bin/bash

# This scripts reads "deployments/versions" and outputs the desired
# version for each tool specified in the directory to Github Action output
# (or stdout in local).

set -eu

targetDir=${1:-"deployments/versions"}
targetDirAbs="${BASH_SOURCE%/*}/../../../${targetDir}"

for file in "${targetDirAbs}"/*; do
  filename="${file##*/}"
  required_version="$(cat "${file}" || exit 1)"
  echo "${filename}=${required_version}" >> "${GITHUB_OUTPUT:-/dev/stdout}"
done

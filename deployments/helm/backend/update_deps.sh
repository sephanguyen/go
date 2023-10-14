#!/bin/bash

set -euo pipefail

currentdir=$(dirname "$0")
for dir in "${currentdir}"/*; do
  if [ ! -d "${dir}" ]; then
    continue
  fi

  sourceDir="${currentdir}/../libs/util/templates/."
  destDir="${dir}/templates/util"

  echo "Copying ${sourceDir} to ${destDir}"
  rm -rf "${destDir}"
  mkdir -p "${destDir}"
  cp -a "${sourceDir}" "${destDir}"
done

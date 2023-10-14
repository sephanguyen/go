#!/bin/bash

set -euo pipefail

currentdir="${BASH_SOURCE%/*}"
sourceDir="${currentdir}/../../libs/util/templates/."
destDir="${currentdir}/templates/util"

echo "Copying ${sourceDir} to ${destDir}"
rm -rf "${destDir}"
mkdir -p "${destDir}"
cp -a "${sourceDir}" "${destDir}"

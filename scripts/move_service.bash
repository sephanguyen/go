#!/bin/bash
set -euo pipefail
SVC=$1

sourceDir="./deployments/helm/manabie-all-in-one/charts/${SVC}/."
destDir="./deployments/helm/backend/${SVC}"
echo "Copy entire folder ${sourceDir} to ${destDir}"
cp -ar $sourceDir $destDir

echo "Update util"
make update-deps

for valuePath in ./deployments/helm/manabie-all-in-one/*values.yaml;
do
  filename="${valuePath##*/}"
  echo "Adding value in ${filename}"
  if [[ "${filename}" == "values.yaml" ]];
  then
    echo "$(yq ".${SVC}" $valuePath)" >> "${destDir}/${filename}"
  else
    echo "$(yq ".${SVC}" $valuePath)" > "${destDir}/${filename}"
  fi
done

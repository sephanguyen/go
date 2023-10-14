#!/bin/bash

get_required_version() {
  binname="$1"
  cat "${BASH_SOURCE%/*}/../../../deployments/versions/${binname}" || echo "failed-to-get-version"
}

tools=("$@")
toolVersions=()
for tool in "${tools[@]}"; do
  toolVersions+=("${tool}@$(get_required_version "${tool}")")
done

echo "${toolVersions[@]}"

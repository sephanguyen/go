#!/bin/bash

# For preproduction, when we deploy, we must run this script to clone
# the configs/secrets from prod->preprod.

set -euo pipefail

projectRootDir="${BASH_SOURCE%/*}/../.."

# Clone to preprod for platform services
platformServices=(
  "elastic"
  "kafka-connect"
  "nats-jetstream"
  "unleash"
)
configTypes=(
  "configs"
  "connectors" # for kafka
  "secrets"
)
for svc in "${platformServices[@]}"; do
  for configType in "${configTypes[@]}"; do
    sourceDir="${projectRootDir}/deployments/helm/platforms/${svc}/${configType}/${ORG}/prod"
    targetDir="${projectRootDir}/deployments/helm/platforms/${svc}/${configType}/${ORG}/dorp"
    if [ -d "${sourceDir}" ] && [ ! -d "${targetDir}" ]; then
      echo "Cloning ${configType} for service ${svc}"
      cp -r "${sourceDir}/." "${targetDir}"
    fi
  done
done

# Clone to preprod for backend services
clone() {
  dir="$1"
  service=${dir##*/}
  for configType in "${configTypes[@]}"; do
    sourceDir=${dir}/${configType}/${ORG}/prod
    targetDir=${dir}/${configType}/${ORG}/dorp
    if [ -d "${sourceDir}" ] && [ ! -d "${targetDir}" ]; then
      echo "Cloning ${configType} for service ${service}"
      cp -r "${sourceDir}/." "${targetDir}"
    fi
  done
}
configTypes=(
  "configs"
  "secrets"
)
for dir in "${projectRootDir}/deployments/helm/manabie-all-in-one/charts/"*; do
  clone "${dir}"
done
for dir in "${projectRootDir}/deployments/helm/backend/"*; do
  clone "${dir}"
done

# Some configurations are hard-coded values for prod environments.
# E.g.: URLs to elasticsearch clusters
# The following step override those values with dorp environment values.
# Overrides for secrets are unimplemented.
override() {
  dir="$1"
  service=${dir##*/}
  sourceDir=${dir}/configs/${ORG}/prod
  targetDir=${dir}/configs/${ORG}/dorp

  # TODO: we can intelligently iterate through all the files in the directory instead
  configFiles=(
    "${service}_migrate.config.yaml"
    "${service}.config.yaml"
    "hasura.config.yaml"
  )
  for file in "${configFiles[@]}"; do
    sourceFile=${sourceDir}/${file}
    targetFile=${targetDir}/${file}
    overrideFile=${targetDir}/override.${file}
    if [ -f "${sourceFile}" ] && [ -f "${overrideFile}" ]; then
      echo "Merging override.${file} into ${file} for service ${service}"
      # Reference: https://mikefarah.gitbook.io/yq/operators/multiply-merge#merge-two-files-together
      yq ". *= load(\"${overrideFile}\")" "${sourceFile}" > "${targetFile}"
    fi
  done
}
for dir in "${projectRootDir}/deployments/helm/manabie-all-in-one/charts/"*; do
  override "${dir}"
done
for dir in "${projectRootDir}/deployments/helm/backend/"*; do
  override "${dir}"
done

#!/bin/bash

set -eux

env=$1
org=$2

dir="deployments/helm/platforms/nats-jetstream/secrets/${org}/${env}"
for file in "${dir}"/*; do
  filename=$(basename "${file}")
  if [[ "${filename}" == *v2* ]]; then
    rm "${file}"
    continue
  fi
  if [[ "${filename}" == "controller"* ]]; then
    continue
  fi
  newfilename="${filename%%\.*}.secrets.encrypted.env"
  data=$(sops -d "${dir}/${filename}" | yq '.data' | yq)
  key=$(echo "${data}" | yq 'keys | .[0]')
  value=$(echo "${data}" | yq ."${key}")
  echo "${key}=\"${value}\"" > "${dir}/${newfilename}"
  sops -i -e "${dir}/${newfilename}"
done

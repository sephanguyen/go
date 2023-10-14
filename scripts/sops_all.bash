#!/bin/bash

set -euo pipefail

action=${ACTION:-d}

doSops() {
  svc=$1
  org=$2
  env=$3
  secretfile="deployments/helm/manabie-all-in-one/charts/${svc}/secrets/${org}/${env}/${svc}_migrate.secrets.encrypted.yaml"
  if [ -f "${secretfile}" ]; then
    if [[ "${action}" == "d" ]]; then
      echo "Decrypting ${secretfile}"
      sops -i -d "${secretfile}"
    else
      echo "Encrypting ${secretfile}"
      sops -i -e "${secretfile}" || echo "failed to encrypt"
    fi
  else
    echo "File in $svc, $org, $env does not exist, skipping..."
  fi
}

svc=$1
orgList=("aic" "ga" "jprep" "manabie" "renseikai" "synersia" "tokyo")
envList=("local" "stag" "uat" "prod")
if [[ "${action}" == "d" ]]; then
  echo "decrypting all secrets in service ${svc}...."
else
  echo "decrypting all secrets in service ${svc}...."
fi

for env in "${envList[@]}"; do
  for org in "${orgList[@]}"; do
    doSops "$svc" "$org" "$env"
  done
done

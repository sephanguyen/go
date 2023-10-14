#!/bin/bash

set -eu

if ! command -v jq >/dev/null; then
  apt-get update && apt-get install -y jq
fi
if ! command -v yq >/dev/null; then
  apt-get update && apt-get install -y curl
  curl -o /usr/local/bin/yq -L https://github.com/mikefarah/yq/releases/download/v4.34.1/yq_linux_amd64
  chmod +x /usr/local/bin/yq
fi

export RESOURCE=${RESOURCE:-}
export CI=true
export NAMESPACE="$ENV-$ORG-backend"
export DATABASE=${DATABASE:-}

# get configuration variable from template
envObj=$(jq -r .$ENV ./deployments/configuration-template.json)
orgObj=$(echo $envObj | jq -r .$ORG)
export PROJECT=$(echo $orgObj | jq -r '.projectId')
export REGION=$(echo $orgObj | jq -r '.region')
export CLUSTER=$(echo $orgObj | jq -r '.cluster')
export DB_PREFIX=$(echo $orgObj | jq -r '.dbPrefix')

sqlProxyConnectionName=$(echo $orgObj | jq -r '.sqlProxyConnectionName')
if [[ "$RESOURCE" == "sql" && ! -z "$DATABASE" ]]; then 
  sqlProxyDatabase=$(echo $orgObj | jq -r .$DATABASE)
  if [[ "$sqlProxyDatabase" != "null" ]]; then
    sqlProxyConnectionName=$sqlProxyDatabase
  fi
fi
export SQL_PROXY_CONN_NAME=${sqlProxyConnectionName}

if [[ "$RESOURCE" == "k8s" ]]; then
  # ad-hoc-k8s
  ./scripts/setup_kubectl.bash
  ./scripts/ad-hoc/execute-k8s.bash
else
  # ad-hoc-sql
  bash "${BASH_SOURCE%/*}/../tests/check_postgres_connection.sh"
  ./scripts/ad-hoc/execute-sql.bash
fi

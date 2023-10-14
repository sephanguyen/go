#!/bin/bash

# This script deploys j4 to an existing k8s cluster.

set -euo pipefail

export PROJECT_ID=local
export ENV=local
export ORG=${ORG:-manabie}
export NAMESPACE=${NAMESPACE:-backend}
export J4_REPLICAS=${J4_REPLICAS:-3}
export SQL_PROXY_CONN_NAME=${SQL_PROXY_CONN_NAME:-"staging-manabie-online:asia-southeast1:manabie-59fd"}
export SCENARIO_NAME=${SCENARIO_NAME:-tom}
export J4_TAG=${J4_TAG:-locally}
export J4_SINGLE_DEPLOYMENT=${J4_SINGLE_DEPLOYMENT:-true}
export J4_DNS_BOOT=${J4_DNS_BOOT:-true}

skaffold run -f ./skaffold.j4.yaml

#!/bin/bash

# This scripts:
# - takes in BE_TAG, FE_TAG, ME_TAG
# - if any of them is empty, fetches the current tag from kubernetes
# - outputs the finalized values

set -euox pipefail  # set -x temporarily for tracing

ENV=${ENV:-local}
ORG=${ORG:-manabie}
GITHUB_OUTPUT=${GITHUB_OUTPUT:-/dev/stdout}
# Always remember to normalize ENV variable on github action.
if [[ "$ENV" == "production" ]]; then
  ENV="prod"
elif [[ "$ENV" == "staging" ]]; then
  ENV="stag"
fi

if [[ "$ENV" == "local" ]]; then
  BE_NAMESPACE="backend"
else
  BE_NAMESPACE="$ENV-$ORG-backend"
fi

FE_NAMESPACE="$ENV-$ORG-frontend"


# Example:
# - Current tag on k8s:
#     BE_TAG=20220909014752.4173e60
#     FE_TAG=20220909035400.b47ac7e-manabie-uat
#     ME_TAG=20220909022637.d5ee3cc-manabie-learner-uat
# -> Function returns
#     20220909014752.4173e60
#     20220909035400.b47ac7e
#     20220909022637.d5ee3cc
function get_current_tag() {
  local image
  local exitcode=0
  local service=$1
  local namespace=$2

  image=$(kubectl get deploy/"$service" -o yaml -n "$namespace" --output=jsonpath='{.spec.template.spec.containers[0].image}') || exitcode=$?
  if [[ ${exitcode} -ne 0 ]]; then
    if [ "${service}" != "bob" ] && [ "${ORG}" == "aic" ]; then
      return 0
    fi
    
    if [ "${service}" != "backoffice" ] && [ "${ORG}" == "aic" ]; then
      return 0
    fi
    if [ "${service}" != "backoffice" ] && [ "${ORG}" == "synersia" ]; then
      return 0
    fi
    
    
    if [ "${service}" != "learner" ] && [ "${ORG}" == "aic" ]; then
      return 0
    fi
    if [ "${service}" != "learner" ] && [ "${ORG}" == "synersia" ]; then
      return 0
    fi
    

    return 1
  fi
  local tag=${image##*:} # trim image name component
  tag=${tag%%-*}  # trim env/org components
  echo "$tag"
}

BE_TAG=${BE_TAG:-"$(get_current_tag bob ${BE_NAMESPACE})"}
FE_TAG=${FE_TAG:-"$(get_current_tag backoffice ${FE_NAMESPACE})"}
ME_TAG=${ME_TAG:-"$(get_current_tag learner ${FE_NAMESPACE})"}

echo "BE_TAG: ${BE_TAG}"
echo "FE_TAG: ${FE_TAG}"
echo "ME_TAG: ${ME_TAG}"
{ echo "BE_TAG=${BE_TAG}";
echo "FE_TAG=${FE_TAG}";
echo "ME_TAG=${ME_TAG}"; } >> $GITHUB_OUTPUT

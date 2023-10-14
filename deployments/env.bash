#!/bin/bash

set -euo pipefail

# shellcheck source=../scripts/log.sh
source ./scripts/log.sh

# setup_env_variables adjust the variables depending on whether this is local or CI environment.
# Values that are already set won't be changed.
setup_env_variables() {
  if [[ -z "$CI" ]]; then
    echo "Cannot determine environment variables: CI is not set"
    exit 1
  fi

  check_deprecated "KAFKA_DEPLOYMENT_ENABLED"
  check_deprecated "IMD_DEPLOYMENT_ENABLED"
  check_deprecated "BACKOFFICE_DEPLOYMENT_ENABLED"
  check_deprecated "LEARNER_DEPLOYMENT_ENABLED"
  check_deprecated "TEACHER_DEPLOYMENT_ENABLED"
  check_deprecated "HASURA_DEPLOYMENT_ENABLED" 'If you want to disable Hasura, try using "-p no-hasura" argument instead. For example: "./deployments/sk.bash -- -p no-hasura"'
  check_deprecated "ELASTIC_DEPLOYMENT_ENABLED"
  check_deprecated "UNLEASH_DEPLOYMENT_ENABLED"
  check_deprecated "YUGABYTE_DEPLOYMENT_ENABLED"
  check_deprecated "USE_KIND"

  env="local"
  org="manabie"
  server_image_tag=locally
  aphelios_deployment_enabled=false
  redash_deployment_enabled=false
  appsmith_deployment_enabled=false
  disable_gateway=false
  camel_k_enabled=false
  network_policy_enabled=false
  local_registry_domain="localhost:5001"
  if [[ "${USE_SHARED_REGISTRY:-false}" == "true" ]]; then
    local_registry_domain="kind-reg.actions-runner-system.svc"
  fi
  artifact_registry_domain="localhost:5001"
  if [[ "${USE_SHARED_REGISTRY:-false}" == "true" ]]; then
    artifact_registry_domain="asia-southeast1-docker.pkg.dev/student-coach-e1e95/ci"
  fi

  case "$CI" in
  false)
    cpu_limit=4
    memory_limit=12g
    ;;
  true)
    cpu_limit=max
    memory_limit=max
    ;;
  *)
    echo "Invalid \"CI\" value \"$CI\", must be one of: true, false"
    ;;
  esac

  export DECRYPT_KEY=9ef85f8fcde4139b88bbbfe5
  export CPU_LIMIT=${CPU_LIMIT:-${cpu_limit}}
  export MEMORY_LIMIT=${MEMORY_LIMIT:-${memory_limit}}

  # environment variables for local env
  export ENV=${ENV:-${env}}
  export ORG=${ORG:-${org}}
  export TAG=${TAG:-${server_image_tag}}
  export BACKOFFICE_TAG=${BACKOFFICE_TAG:-locally}
  export LEARNER_TAG=${LEARNER_TAG:-locally}
  export TEACHER_TAG=${TEACHER_TAG:-locally}
  export NAMESPACE=${NAMESPACE:-backend}
  export INSTALL_MONITORING_STACKS=${INSTALL_MONITORING_STACKS:-false}
  export APHELIOS_DEPLOYMENT_ENABLED=${APHELIOS_DEPLOYMENT_ENABLED:-${aphelios_deployment_enabled}}
  export REDASH_DEPLOYMENT_ENABLED=${REDASH_DEPLOYMENT_ENABLED:-${redash_deployment_enabled}}
  export CAMEL_K_ENABLED=${CAMEL_K_ENABLED:-${camel_k_enabled}}
  export NETWORK_POLICY_ENABLED=${NETWORK_POLICY_ENABLED:-${network_policy_enabled}}
  export SERVICE_ACCOUNT_EMAIL_SUFFIX=""
  export DISABLE_GATEWAY=${DISABLE_GATEWAY:-${disable_gateway}}
  export LOCAL_REGISTRY_DOMAIN=${LOCAL_REGISTRY_DOMAIN:-${local_registry_domain}}
  export ARTIFACT_REGISTRY_DOMAIN=${ARTIFACT_REGISTRY_DOMAIN:-${artifact_registry_domain}}
  export APPSMITH_DEPLOYMENT_ENABLED=${APPSMITH_DEPLOYMENT_ENABLED:-${appsmith_deployment_enabled}}

  # skaffold-related settings
  export SKAFFOLD_UPDATE_CHECK=false

  if [[ "$ENV" == "local" ]]; then
    SVC_CRED=$(gpg --quiet --batch --yes --decrypt --passphrase="$DECRYPT_KEY" ./developments/serviceaccount.json.base64.gpg)
  fi
  export SVC_CRED=${SVC_CRED:-}

  # Print out the values when they are overriden
  if [[ "${CI}" != "false" ]]; then
    echo "** CI=\"${CI}\""
  fi
  if [[ "${CPU_LIMIT}" != "${cpu_limit}" ]]; then
    echo "** CPU_LIMIT=\"${CPU_LIMIT}\""
  fi
  if [[ "${MEMORY_LIMIT}" != "${memory_limit}" ]]; then
    echo "** MEMORY_LIMIT=\"${MEMORY_LIMIT}\""
  fi
  if [[ "${ENV}" != "${env}" ]]; then
    echo "** ENV=\"${ENV}\""
  fi
  if [[ "${ORG}" != "${org}" ]]; then
    echo "** ORG=\"${ORG}\""
  fi
  if [[ "${TAG}" != "${server_image_tag}" ]]; then
    echo "** TAG=\"${TAG}\""
  fi
  if [[ "${INSTALL_MONITORING_STACKS}" != "false" ]]; then
    echo "** INSTALL_MONITORING_STACKS=\"${INSTALL_MONITORING_STACKS}\""
  fi
  if [[ "${SERVICE_ACCOUNT_EMAIL_SUFFIX}" != "" ]]; then
    echo "** SERVICE_ACCOUNT_EMAIL_SUFFIX=\"${SERVICE_ACCOUNT_EMAIL_SUFFIX}\""
  fi
  if [[ "${APHELIOS_DEPLOYMENT_ENABLED}" != "${aphelios_deployment_enabled}" ]]; then
    echo "** APHELIOS_DEPLOYMENT_ENABLED=\"${APHELIOS_DEPLOYMENT_ENABLED}\""
  fi
  if [[ "${REDASH_DEPLOYMENT_ENABLED}" != "${redash_deployment_enabled}" ]]; then
    echo "** REDASH_DEPLOYMENT_ENABLED=\"${REDASH_DEPLOYMENT_ENABLED}\""
  fi
  if [[ "${CAMEL_K_ENABLED}" != "${camel_k_enabled}" ]]; then
    echo "** CAMEL_K_ENABLED=\"${CAMEL_K_ENABLED}\""
  fi
  if [[ "${NETWORK_POLICY_ENABLED}" != "${network_policy_enabled}" ]]; then
    echo "** NETWORK_POLICY_ENABLED=\"${NETWORK_POLICY_ENABLED}\""
  fi
  if [[ "${DISABLE_GATEWAY}" != "${disable_gateway}" ]]; then
    echo "** DISABLE_GATEWAY=\"${DISABLE_GATEWAY}\""
  fi
  if [[ "${LOCAL_REGISTRY_DOMAIN}" != "${local_registry_domain}" ]]; then
    echo "** LOCAL_REGISTRY_DOMAIN=\"${LOCAL_REGISTRY_DOMAIN}\""
  fi
  if [[ "${ARTIFACT_REGISTRY_DOMAIN}" != "${artifact_registry_domain}" ]]; then
    echo "** ARTIFACT_REGISTRY_DOMAIN=\"${ARTIFACT_REGISTRY_DOMAIN}\""
  fi
  if [[ "${APPSMITH_DEPLOYMENT_ENABLED}" != "${appsmith_deployment_enabled}" ]]; then
    echo "** APPSMITH_DEPLOYMENT_ENABLED=\"${APPSMITH_DEPLOYMENT_ENABLED}\""
  fi
}

function check_deprecated() {
  var="$1"
  if [ -n "${!var+x}" ]; then
    msg="(deprecation) env \"$var\" is set to \"${!var}\", but will be ignored."
    if [ -n "${2+x}" ]; then
      msg="${msg}\n${2}"
    fi
    logwarn "${msg}"
  fi
}

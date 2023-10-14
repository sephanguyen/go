#!/bin/bash

# This script copies the latest binaries built in local machine
# to the remote pods running inside Kubernetes (as part of the 
# live reload process).

set -euo pipefail

readonly SKAFFOLD_FILE=${SKAFFOLD_FILE:-"skaffold2.local.yaml"}
readonly KIND_CONTEXT=${KIND_CONTEXT:-"kind-kind"}
readonly BACKEND_NAMESPACE=${BACKEND_NAMESPACE:-"backend"}
SVC_LIST=()
BUILD_SERVER=${BUILD_SERVER:-false}
BUILD_BDD_TEST=${BUILD_BDD_TEST:-false}
BUILD_STUB=${BUILD_STUB:-false}
kubecontext=${KIND_CONTEXT}

cmdname=${0##*/}
fullcmdname="./deployments/${cmdname}"
function usage() {
  cat <<EOF
Live-reload backend servers by:
  - build go code
  - copy the binaries to running containers in Kubernetes
  - the server process then will be reloaded with modd

!! Warning: this script only reloads the go binary. To apply changes
from non-go code (configs, migrations, feature files, ...),
run "deployments/sk.bash -f skaffold2.backend.yaml" instead.

Examples:
  # Build and reload some services.
  # Should be used when server code is updated.
  ${fullcmdname} bob,tom,yasuo

  # Use "gandalf" service to reload gandalf-ci and/or gandalf-stub.
  # Should be used when integration test code is updated.
  ${fullcmdname} gandalf
  ${fullcmdname} gandalf-ci   # skip gandalf-stub

Options:
  ${cmdname} [options]
    -h, --help: Print help.

Usage:
  ${fullcmdname} SVCNAME[,SVCNAME,[...]]
EOF
}

# shellcheck source=../scripts/log.sh
source ./scripts/log.sh

# process arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    -h | --help)
      usage
      exit 0
      ;;
    *)
      if [[ "$1" == "-"* ]]; then
        # only accept non-flag args
        logfatal "Invalid flag: $1"
      fi
      SVC_LIST+=("$1")
      shift 1
  esac
done

readonly SVC_LIST
if [[ ${#SVC_LIST[@]} -eq 0 ]]; then
  logfatal "At least 1 service must be specified. See \"${fullcmdname} --help\" for more information."
fi

function build() {
  build_cmd="CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v"
  if [[ "${BUILD_SERVER}" == "true" ]]; then
    loginfo "Building server"
    eval "$build_cmd" -o ./build/server ./cmd/server/
  fi
  if [[ "${BUILD_BDD_TEST}" == "true" ]]; then
    loginfo "Building bdd.test"
    eval "$build_cmd" -o ./build/bdd.test ./features/
  fi
  if [[ "${BUILD_STUB}" == "true" ]]; then
    loginfo "Building gandalf stub"
    eval "$build_cmd" -o ./build/stub ./features/stub/
  fi
}

function inject_binary() {
  local svc="$1"
  local backend_namespace="${BACKEND_NAMESPACE}"
  local helm_chart="manabie-all-in-one"
  local svc_type="deployment"
  if [[ "${svc}" == "tom" ]]; then
    svc_type="statefulset"
  fi

  if ! kubectl -n "${backend_namespace}" get "${svc_type}" "${svc}" &> /dev/null; then
    backend_namespace="local-manabie-backend"
    if [[ "${svc}" == "gandalf-ci" || "${svc}" == "gandalf-stub" ]]; then
      helm_chart="gandalf"
    else
      helm_chart="${svc}"
    fi
  fi
  loginfo "Service assumed to be in \"${backend_namespace}\" namespace"

  local pods=()
  mapfile -t pods < <(
    kubectl get pods \
      --context="${kubecontext}" \
      -n "${backend_namespace}" \
      --no-headers \
      --selector=app.kubernetes.io/instance="${helm_chart%-caching*}",app.kubernetes.io/name="${svc}" \
      --output=custom-columns=":metadata.name" \
    )

  if [[ ${#pods[@]} -eq 0 ]]; then
    logwarn "Unable to find any pods for \"${svc}\" service"
    return 0
  fi

  local kcp_cmd="kubectl -n ${backend_namespace} --context=${kubecontext} cp"
  for p in "${pods[@]}"; do
    if [[ "${svc}" == "gandalf-ci" ]]; then
      loginfo "Copying ./features into ${p}/${svc}"
      $kcp_cmd "./features" "$p":/backend/ -c "${svc}"
      loginfo "Copying ./build/bdd.test into ${p}/${svc}"
      $kcp_cmd "./build/bdd.test" "$p":/backend/features/bdd.test -c "${svc}"
    elif [[ "${svc}" == "gandalf-stub" ]]; then
      loginfo "Copying ./build/stub into ${p}/${svc}"
      $kcp_cmd "./build/stub" "$p":/stub -c "${svc}"
    else
      loginfo "Copying ./build/server into ${p}/${svc}"
      $kcp_cmd "./build/server" "$p":/server -c "${svc}"
    fi
  done
}

for s in "${SVC_LIST[@]}"; do
  if [[ "${s}" == "gandalf" ]]; then
    BUILD_BDD_TEST=true
    BUILD_STUB=true
  elif [[ "${s}" == "gandalf-ci" ]]; then
    BUILD_BDD_TEST=true
  elif [[ "${s}" == "gandalf-stub" ]]; then
    BUILD_STUB=true
  else
    BUILD_SERVER=true
  fi
done

readonly BUILD_SERVER
readonly BUILD_BDD_TEST
readonly BUILD_STUB

build

# Copy all built binaries/required files into docker
for s in "${SVC_LIST[@]}"; do
  # Some exceptions:
  # - gandalf is sugar opt for gandalf-ci + gandalf-stub
  if [[ "${s}" == "gandalf" ]]; then
    inject_binary "gandalf-ci"
    inject_binary "gandalf-stub"

  # - eureka has additional deployments
  elif [[ "${s}" == "eureka" ]]; then
    inject_binary "eureka"
    inject_binary "eureka-all-consumers"
    inject_binary "eureka-monitors"

  # - for others: single deployment/statefulset
  else
    inject_binary "${s}"
  fi
done

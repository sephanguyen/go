#!/bin/bash

set -euo pipefail

SECONDS=0 # measuring execution time
SKAFFOLD_FILE=${SKAFFOLD_FILE:-"skaffold2.local.yaml"}
FLAG_BUILD=${FLAG_BUILD:-""}
FLAG_INSTALL=${FLAG_INSTALL:-""}
FLAG_SERVICES=${FLAG_SERVICES:-""}
FLAG_DELETE=${FLAG_DELETE:-"false"}
FLAG_RAW=()
use_shared_registry="${USE_SHARED_REGISTRY:-false}"
MANABIE_DEPLOYER_ENABLED=${MANABIE_DEPLOYER_ENABLED:-false}
readonly KIND_CONTEXT=${KIND_CONTEXT:-"kind-kind"}
readonly BACKEND_NAMESPACE=${BACKEND_NAMESPACE:-"backend"}

kubecontext=${KIND_CONTEXT}

# shellcheck source=../scripts/log.sh
source ./scripts/log.sh

# use shared registry or not
if [[ "${use_shared_registry}" == "true" ]]; then
    echo "Using shared registry"
    ./deployments/prepare_shared_registry.bash
fi

function collect_run_time {
  loginfo "Deployment took: ${SECONDS} seconds"
}

cmdname=${0##*/}
fullcmdname="./deployments/${cmdname}"
function usage() {
  cat <<EOF
Run backend in a local Kubernetes cluster.

Examples:
  # Deploy (or redeploy) everything
  ${fullcmdname}

  # Rebuild and redeploy certain backend services
  ${fullcmdname} -s bob,tom

  # Rebuild certain backend services (without restarting pods)
  ${fullcmdname} -b bob,tom

  # Rebuild and redeploy certain services
  ${fullcmdname} -i bob,tom

  # Delete everything
  ${fullcmdname} -d

  # Arguments after \`--\` are passed as-is to the underlying \`skaffold\` command
  # If no command is specify, default is \`skaffold run\`
  ${fullcmdname} -- -p local          # skaffold run with local profile
  ${fullcmdname} -- run -p prod       # skaffold run with prod profile
  ${fullcmdname} -- diagnose          # skaffold diagnose

Options:
  ${cmdname} [options]
    -b, --build: If specified, build for the specified backend services. To build with deployment, use -i.
    -d, --delete: Delete local cluster entirely.
    -f, --file: Skaffold file to use (default: ${SKAFFOLD_FILE}).
    -h, --help: Print help.
    -i, --install: If specified, build and redeploy for the specified backend services. To build without deployment, use -b.
    -s, --services: If specified, build and redeploy for only the specified backend services (default: [${FLAG_SERVICES}]).
EOF
}

# process arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    -b | --build)
      FLAG_BUILD=$2
      if [[ -z "$FLAG_BUILD" ]]; then
        logwarn "list of services is empty. -b/--build will therefore be ignored"
      fi
      shift 2
      ;;
    -h | --help)
      usage
      exit 0
      ;;
    -d | --delete)
      FLAG_DELETE="true"
      shift 1
      ;;
    -f | --file)
      SKAFFOLD_FILE=$2
      shift 2
      ;;
    -i | --install)
      logfatal "-i/--install not yet implemented"
      FLAG_INSTALL=$2
      if [[ -z "$FLAG_INSTALL" ]]; then
        logwarn "list of services is empty. -i/--install will therefore be ignored"
      fi
      shift 2
      ;;
    -s | --services)
      FLAG_SERVICES=$2
      if [[ -z "$FLAG_SERVICES" ]]; then
        logwarn "list of services is empty. -s or --services will therefore be ignored"
      fi
      shift 2
      ;;
    --)
      # Stop parsing commands from here
      shift 1
      FLAG_RAW=("$@")
      break
      ;;
    *)
      logfatal "Unrecognized argument: $1
See \"${fullcmdname} --help\" for more information."
  esac
done

readonly FLAG_SERVICES
readonly FLAG_RAW

########## Various functions handling different cases for this script

# Start the local k8s cluster. Use delete_cluster to delete it.
function start_cluster() {
  echo "Using kind to deploying on local"
  ./deployments/kind_with_registry.bash
}

# Delete k8s cluster in local. Use start_cluster to start it again.
function delete_cluster() {
  ./scripts/clean_docker_daemon.bash
}

# Restart a service
# Example: restart_service bob
function restart_service() {
  local svc="$1"
  local namespace="${BACKEND_NAMESPACE}"
  svcValueFile="./deployments/helm/backend/${svc}/values.yaml"
  if [[ -f "${svcValueFile}" ]]; then
    envOrgValueFile="./deployments/helm/backend/${svc}/${ENV}-${ORG}-values.yaml"
    if [[ "$(yq '.enabled' ${svcValueFile})" == "true" ]]; then
      if [[ "$(yq '.enabled' ${envOrgValueFile})" != "false" ]]; then
        namespace="${ENV}-${ORG}-backend"
      fi
    else
      if [[ "$(yq '.enabled' "${envOrgValueFile}")" == "true" ]]; then
        namespace="${ENV}-${ORG}-backend"
      fi
    fi
  fi
  restart_cmd="kubectl -n ${namespace} --context=${kubecontext} rollout restart"
  
  if [[ "${svc}" == "gandalf" ]]; then
    $restart_cmd deploy/gandalf-ci deploy/gandalf-stub
  else
    if [[ "$svc" == "tom" ]]; then
      $restart_cmd statefulset/"${svc}"
    else
      $restart_cmd deploy/"${svc}"
    fi
  fi
}

# Rebuild docker image and restart the specified services
# Example: build_with_restart bob,tom,eureka
function build_with_restart() {
  if [[ "${#FLAG_RAW[@]}" -gt 0 ]]; then
    logerror "extra args after -- not allowed for -s option"
    exit 1
  fi
  loginfo "Running: skaffoldv2 run --filename=skaffold2.backend.yaml"
  skaffoldv2 run --filename=skaffold2.backend.yaml
  local svcs="$1"
  for s in ${svcs//,/ }; do
    restart_service "${s}"
  done
}

########## main
export CI=${CI:-false}
if [[ "$CI" == "false" ]]; then
  ./deployments/tools.bash # install skaffold of a specific version
fi
# shellcheck source=env.bash
. ./deployments/env.bash
setup_env_variables


# If -d/--delete is specified, run delete
if [[ "$FLAG_DELETE" == "true" ]]; then
  delete_cluster
  exit 0
fi

# If -b/--build is specified, build and inject those services then exit
if [[ -n "$FLAG_BUILD" ]]; then
  reload_cmd="./deployments/reload.bash ${FLAG_BUILD//,/ }"
  loginfo "Reloading with: ${reload_cmd}"
  eval "${reload_cmd}"
  exit 0
fi

# If -s/--services is specified, build and restart those services and exit
if [[ -n "$FLAG_SERVICES" ]]; then
  build_with_restart "$FLAG_SERVICES"
  exit 0
fi

# Time the script when deploying
trap collect_run_time EXIT

# Parse FLAG_RAW to determine which skaffold command we have to run
skcmd="run"
all_args=()
if [[ "${#FLAG_RAW[@]}" -gt 0 ]]; then
  loginfo "Received raw args: ${FLAG_RAW[*]}"
  firstcmd=${FLAG_RAW[0]}
  if [[ "$firstcmd" == "-"* ]]; then  # check if command or flags
    all_args+=("${FLAG_RAW[@]}")
  else
    skcmd="$firstcmd"
    all_args+=("${FLAG_RAW[@]:1}")
  fi

  # Check for -f and update SKAFFOLD_FILE if need be
  fflag=$( (echo "${FLAG_RAW[@]}" | grep -oE -- "-f skaffold.+yaml" | sed 's/^-f //') || echo "" )
  if [[ "${fflag}" == "skaffold"* ]]; then
    loginfo "Found \"-f\" argument in raw args: ${fflag}"
    SKAFFOLD_FILE="${fflag}"
  fi
fi
readonly SKAFFOLD_FILE

bin_name="skaffold"
if [[ "${SKAFFOLD_FILE}" == "skaffold2"* ]]; then
  bin_name="skaffoldv2"
fi

if [[ "$skcmd" == "run" ]]; then
  all_args=("--default-repo=asia.gcr.io/student-coach-e1e95" "--skip-tests" "${all_args[@]}")
  if [[ "${bin_name}" == "skaffoldv2" ]]; then
    all_args=("--cleanup=false" "${all_args[@]}")
  else
    all_args=("--status-check=false" "${all_args[@]}")
  fi
fi
all_args=("--filename=${SKAFFOLD_FILE}" "${all_args[@]}")
fullcmd=("${bin_name}" "${skcmd}" "${all_args[@]}")

# Also start minikube/kind in a `skaffold run`
if [[ "$skcmd" == "run" ]]; then
  start_cluster
  ./deployments/cache_images.bash
fi

loginfo "Full command: ${fullcmd[*]}"
exitcode=0
if [[ "${MANABIE_DEPLOYER_ENABLED}" == "true" ]]; then
  ./deployer run -v debug || exitcode=$?
else
  "${fullcmd[@]}" || exitcode=$?
fi
if [[ $exitcode != 0 ]]; then
  logerror "helm install returned ${exitcode}"
  if [[ "${skcmd}" == "run" || "${skcmd}" == "deploy" ]]; then
    loginfo "Running diagnostics..."
    ./scripts/diagnose.bash
  fi
  exit ${exitcode}
fi

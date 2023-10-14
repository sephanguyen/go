#!/bin/bash

set -euo pipefail

FLAG_RAW=()
JOB_NAME=""
JOB_SERVICE=""
DELETE_EXISTING_JOB=${DELETE_EXISTING_JOB:-false}
NAMESPACE=${NAMESPACE:-"backend"}
DRY_RUN=${DRY_RUN:-false}
IS_NEW_NAMESPACE=false

readonly KIND_CONTEXT=${KIND_CONTEXT:-"kind-kind"}

cmdname=${0##*/}
fullcmdname="./deployments/${cmdname}"

function usage() {
  cat <<EOF
Usage: ${fullcmdname} [flags] SERVICE JOB [-- JOB_FLAGS...]

Deploy a job in your current Kubernetes cluster.
To run a job on stag/uat/prod cluster, use the "deploy-job" Github Workflow instead.

Note: the job will use the image currently deployed in the cluster. Therefore, deploy first before triggering jobs.

Examples:
  # Run job "test-job" defined in eureka/values.yaml
  ${fullcmdname} eureka test-job

  # Remove the job before running it
  ${fullcmdname} eureka test-job -d
  
  # Passing additional flags to job
  # Effectively, the command below will run: /server eureka test-job --testID=1234 --testName="Elon Musk"
  ${fullcmdname} eureka test-job -- --testID=1234 --testName="Elon Musk"

Creating jobs:
  To create a new job, add a new entry in manabie-all-in-one/charts/<service>/values.yaml.
  For example:

  # eureka/values.yaml
    jobs:
      test-job:               # name of job
        cmd: eureka_test_job  # required, command to run
        enabled: false        # optional, must always be "false", as we
                              #   must always trigger job manually
                              #   setting to "true" will cause helm
                              #   installation to fail (not for the first,
                              #   but for the ones after)
        args:                 # optional, default flag values (if any)
          flag1: "default 1"
          flag2: "..."

  Running with the above configuration would create "eureka-test-job" running this command:
    /server gjob \\
        eureka_test_job \\
        --commonConfigPath=/configs/eureka.common.config.yaml \\
        --configPath=/configs/eureka.config.yaml \\
        --secretsPath=/configs/eureka.secrets.encrypted.yaml \\
        --someKey='some default value' \\
        ;

Options:  
    -d, --delete: Delete existing job (default: false)
    -h, --help: Print help.
        --dry-run: Print manifests without applying
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
    -d | --delete)
      DELETE_EXISTING_JOB="true"
      shift 1
      ;;
    --dry-run)
      DRY_RUN="true"
      shift 1
      ;;
    --)
      # Stop parsing commands from here
      shift 1
      FLAG_RAW=("$@")
      break
      ;;
    *)
      if [[ -z "$JOB_SERVICE" ]]; then
        readonly JOB_SERVICE=$1
      elif [[ -z "$JOB_NAME" ]]; then
        readonly JOB_NAME=$1
      else
        logerror "Got more than 2 arguments, encountered: $1"
        usage
        exit 1
      fi
      shift 1
      ;;
  esac
done

readonly FLAG_RAW

if [[ -z "$JOB_SERVICE" ]]; then
  logfatal "no job service specified"
fi
if [[ -z "$JOB_NAME" ]]; then
  logfatal "no job name specified"
fi

export CI=${CI:-false}
# shellcheck source=env.bash
. ./deployments/env.bash
setup_env_variables
# Setup variables for CI or local
kubecontext=""
vpaEnabled="true"

if [[ "$CI" == "false" ]]; then
  kubecontext=${KIND_CONTEXT}
  vpaEnabled="false"
fi
readonly kubecontext

serviceKind="deploy"
if [[ "${JOB_SERVICE}" == "tom" ]]; then
  serviceKind="statefulset"
fi

exitcode=0
kubectl get "${serviceKind}"/"${JOB_SERVICE}" -n "${NAMESPACE}" || exitcode=$?
svcValueFile="./deployments/helm/manabie-all-in-one/charts/${JOB_SERVICE}/values.yaml"

if [[ ${exitcode} -ne 0 ]]; then
  svcValueFile="./deployments/helm/backend/${JOB_SERVICE}/values.yaml"
  IS_NEW_NAMESPACE=true
fi

if [[ ! -f "${svcValueFile}" ]]; then
  logfatal "no values file at ${svcValueFile} (is service ${JOB_SERVICE} valid?)"
fi

existsKeyJob=$(yq 'has("jobs")' < "${svcValueFile}")
if [[ "${existsKeyJob}" != "true" ]]; then
  logfatal "no jobs defined in service ${JOB_SERVICE}"
fi

existsKeyJobname=$(yq "... comments=\"\" | .jobs | has(\"${JOB_NAME}\")" < "${svcValueFile}")
if [[ "${existsKeyJobname}" != "true" ]]; then
  logfatal "job ${JOB_NAME} does not exist in service ${JOB_SERVICE}; available are:\n$(yq '... comments="" | .jobs | keys' < "${svcValueFile}")"
fi

if [[ "${IS_NEW_NAMESPACE}" == "true" ]]; then
  NAMESPACE="${ENV}-${ORG}-backend"
fi

# Setup image variables
image=""
if [[ "$JOB_SERVICE" == "tom" ]] ; then
  image=$(kubectl --context="$kubecontext" get sts -n "$NAMESPACE" -o=jsonpath='{$.spec.template.spec.containers[:1].image}' "$JOB_SERVICE")
else
  image=$(kubectl --context="$kubecontext" get deployment -n "$NAMESPACE" -o=jsonpath='{$.spec.template.spec.containers[:1].image}' "$JOB_SERVICE")
fi
imageRepository=${image%:*}
imageTag=${image##*:}

set +u
# Server Args to Helm Set Flags
serverArgs=()
for i in "${!FLAG_RAW[@]}"; do
  flag=${FLAG_RAW[$i]}
  valid=$(echo "$flag" | grep -e '--.*=.*' || echo -n "")
  if [[ "$valid" == "" ]]; then
    logfatal "flag $flag invalid \njob arguments must be of the format --key=value";
  fi

  flag=${flag:2}
  arg=${flag%%=*}
  value=${flag#*=}
  if [[ "${IS_NEW_NAMESPACE}" == "true" ]]; then
    serverArgs+=("--set=jobs.$JOB_NAME.args.$arg=${value}")
  else
    serverArgs+=("--set=$JOB_SERVICE.jobs.$JOB_NAME.args.$arg=${value}")
  fi
done

# Run Job
loginfo "running job $JOB_SERVICE.$JOB_NAME with arguments ${serverArgs[*]}"

tmpFile=$(mktemp --suffix .yml)
trap "rm $tmpFile" EXIT
if [[ "${IS_NEW_NAMESPACE}" == "true" ]]; then
  helm template -n "${NAMESPACE}" --kube-context="${kubecontext}" "${JOB_SERVICE}"  ./deployments/helm/backend/"${JOB_SERVICE}" \
    --values "./deployments/helm/backend/values.yaml" \
    --values "./deployments/helm/backend/${ENV}-${ORG}-values.yaml" \
    --values "./deployments/helm/backend/${JOB_SERVICE}/values.yaml" \
    --values "./deployments/helm/backend/${JOB_SERVICE}/${ENV}-${ORG}-values.yaml" \
    --values "./deployments/helm/platforms/gateway/${ENV}-${ORG}-values.yaml" \
    --set=enabled=true \
    --set=global.image.repository=$imageRepository \
    --set=image.tag=$imageTag \
    --set=image.pullPolicy=IfNotPresent \
    --set=jobs.$JOB_NAME.enabled="true" \
    --set=global.environment="${ENV}" \
    --set=global.vendor="${ORG}" \
    --set=global.appVersion="$imageTag" \
    --set=global.vpa.enabled="$vpaEnabled" \
    "${serverArgs[@]}" | \
      yq e '. | select(.kind | test("^Job"))'| \
      yq e ". | select(.metadata.name == \"${JOB_SERVICE}-${JOB_NAME}\")" > "$tmpFile"
else
  helm template -n "${NAMESPACE}" --kube-context="${kubecontext}" manabie-all-in-one ./deployments/helm/manabie-all-in-one \
  --values "./deployments/helm/manabie-all-in-one/values.yaml" \
  --values "./deployments/helm/manabie-all-in-one/${ENV}-${ORG}-values.yaml" \
  --values "./deployments/helm/platforms/gateway/${ENV}-${ORG}-values.yaml" \
  --set=global.$JOB_SERVICE.enabled=true \
  --set=global.image.repository=$imageRepository \
  --set=$JOB_SERVICE.image.tag=$imageTag \
  --set=$JOB_SERVICE.image.pullPolicy=IfNotPresent \
  --set=$JOB_SERVICE.jobs.$JOB_NAME.enabled="true" \
  --set=global.environment="${ENV}" \
  --set=global.vendor="${ORG}" \
  --set=global.appVersion="$imageTag" \
  --set=global.vpa.enabled="$vpaEnabled" \
  "${serverArgs[@]}" | \
    yq e '. | select(.kind | test("^Job"))'| \
    yq e ". | select(.metadata.name == \"${JOB_SERVICE}-${JOB_NAME}\")" > "$tmpFile"
fi

if ! grep Job "$tmpFile" >/dev/null; then
  logfatal "Job $JOB_NAME not found in $JOB_SERVICE values.yaml. Nothing to run"
fi

cat "$tmpFile"
if [[ "$DRY_RUN" == "true" ]]; then
  loginfo "Dry run. Stopping"
  exit 0
fi

if [[ "$DELETE_EXISTING_JOB" == "true" ]]; then
  kubectl -n "${NAMESPACE}" --context="${kubecontext}" delete job "$JOB_SERVICE-$JOB_NAME" --ignore-not-found
fi

kubectl -n "${NAMESPACE}" --context="${kubecontext}" apply -f "$tmpFile"


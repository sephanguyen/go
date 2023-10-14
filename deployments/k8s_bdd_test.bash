#!/bin/bash

# Usage: ./deployments/k8s_bdd_test.bash features/bob/add_a_class_member.feature

function usage() {
  cat <<EOF
Run bdd test (Scenarios having tag "@wip" and "@quarantined" are ignored)


Examples:
  # Test all 
  ./deployments/k8s_bdd_test.bash

  # Test all feature files in bob
  ./deployments/k8s_bdd_test.bash bob

  # Test all feature files in bob and tom
  ./deployments/k8s_bdd_test.bash bob tom

  # Test all feature files in bob and communication/tag_create.feature
  ./deployments/k8s_bdd_test.bash bob communication/tag_create.feature
EOF
}

if [[ $1 == "-h" || $1 == "--help" ]]; then
  usage
  exit 0
fi

SECONDS=0 # measuring execution time
readonly KIND_CONTEXT=${KIND_CONTEXT:-"kind-kind"}

kubecontext=${KIND_CONTEXT}
godog_tags='~@wip && ~@quarantined'
test_ignored=("common" "helper" "eibanam" "stub" "repository")
validtestfolder=()

for folder in ./features/*/; do
  folder=${folder/#"./features/"/}
  folder=${folder%"/"}
  if [[ ! "${test_ignored[*]}" =~ ${folder} ]]; then
    validtestfolder+=("$folder")
  fi
done

export CI=${CI:-false}

. ./deployments/env.bash
setup_env_variables

# undo set flags from env.bash
set +euo pipefail

# make build-docker-dev
# to rebuild all images when need
# remember to check release.Dockerfile to pre-build test executable
# Dont specify the feature file to run if you want to run all tests in those services

# If run on local, enable tty so that we can send signal from terminal to the process inside the container.
# Cannot be done on Github Action (it's not a terminal).
tty_arg="--tty"
concurrency=10
if [[ "$CI" == "true" ]]; then
  tty_arg=""
  echo "Running on CI, tty mode is disabled"
fi

# check flag quarantine
if [[ "$RUN_QUARANTINE" == "true" ]]; then
  godog_tags='~@wip && @quarantined'
fi
if [[ ! -z "${GODOG_TAGS}" ]]; then
  godog_tags="${godog_tags} && ${GODOG_TAGS}"
fi
test_state=""

trace_enabled=${TRACE_ENABLED:-false}
otel_endpoint=${OTEL_ENDPOINT:-opentelemetry-collector.manabie.io}

pushgateway_endpoint=${PUSHGATEWAY_ENDPOINT:-https://prometheus-pushgateway.staging.manabie.io}
collect_bdd_tests_metrics=${COLLECT_BDD_TESTS_METRICS:-false}

draft_endpoint=${DRAFT_ENDPOINT:-draft:6050}
ci_pull_request_id="$CI_PULL_REQUEST_ID"
ci_actor="$CI_ACTOR"
ci_run_id="$CI_RUN_ID"

extract_svc_and_path() {
  path=$1
  if [[ "$path" == features/* ]]; then
    path=${path#"features/"}
  fi
  declare -n local_svc=$2
  declare -n local_test_path=$3
  declare -n local_should_run=$4
  IFS='/' read -ra parts <<<"$path"
  local_svc=${parts[0]}
  local_test_path="$path"
  local_should_run=true
  # check if service is deployed, only check draft for now
  if [[ "$local_svc" == "draft" ]]; then
    if ! kubectl -n local-manabie-backend --context=${kubecontext} get deployments.app draft >/dev/null 2>&1; then
      echo "Draft is not deployed, skipping!"
      local_should_run=false
    fi
  fi
  # validate file exist
  if [[ "${#parts[@]}" != "1" ]]; then
    file="features/${path%:*}"
    if [ ! -f "$file" ]; then
      echo "$file does not exist"
      exit 1
    fi
  fi
}

run_arg=("${validtestfolder[@]}")
if [[ "$#" != "0" ]]; then
  run_arg=("${@}")
fi
valid_test_paths=()
for arg in "${run_arg[@]}"; do
  extract_svc_and_path "$arg" svc run_path should_run
  if ! $should_run; then
    continue
  fi
  valid_test_paths+=("$svc $run_path")
done

echo "** Running tests for:"
for arg in "${valid_test_paths[@]}"; do
  echo -e "\t${arg#* }"
done

for arg in "${valid_test_paths[@]}"; do
  read -a args <<<"$arg"
  svc="${args[0]}"
  run_path="${args[1]}"

  gandalfPodName=$(kubectl -n local-manabie-backend --context=${kubecontext} get pods --selector=app.kubernetes.io/role="ci" --no-headers -o custom-columns=":metadata.name" --field-selector=status.phase=Running)
  if [ "$gandalfPodName" == "" ]; then
    echo "No running pod found for gandalf to run test for ${svc}"
    test_state=failed
    continue
  fi

  echo "=================== START TESTING $run_path with $gandalfPodName"
  if [[ "$svc" == "gandalf" ]]; then
    concurrency=1
  else
    concurrency=10
  fi

  kubectl -n local-manabie-backend --context=${kubecontext} exec --stdin $tty_arg $gandalfPodName -c gandalf-ci -- sh -c "cd /backend/features/ && ./bdd.test \
    --godog.tags='${godog_tags}'  \
    --godog.format=progress \
    --godog.concurrency=${concurrency} \
    --godog.random \
    -manabie.service=${svc} \
    -manabie.commonConfigPath=/configs/gandalf.common.config.yaml \
    -manabie.configPath=/configs/gandalf.config.yaml \
    -manabie.secretsPath=/configs/gandalf.secrets.encrypted.yaml \
    -manabie.traceEnabled=$trace_enabled \
    -manabie.otelEndpoint=$otel_endpoint \
    -manabie.pushgatewayEndpoint=$pushgateway_endpoint \
    -manabie.collectBDDTestsMetrics=$collect_bdd_tests_metrics \
    -manabie.draftEndpoint=$draft_endpoint \
    -manabie.ciPullRequestID="$ci_pull_request_id" \
    -manabie.ciActor="$ci_actor" \
    -manabie.ciRunID="$ci_run_id" \
    -- ${run_path}
  "

  if [[ $? > 0 ]]; then
    test_state=failed
    if [[ "$CI" == "true" ]]; then
      echo "Some tests failed. Showing all pods status"
      kubectl get pods --all-namespaces --context=${kubecontext}
    fi
  fi
  echo "=================== END TESTING $svc"
done

if [[ "$test_state" == failed ]]; then
  echo "ERROR: some integration tests failed"
  exit 1
fi

echo "Integration tests completed successfully"
echo "Integration test took: ${SECONDS} seconds"

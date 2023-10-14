#!/bin/bash

# Usage for specific feature
# ./deployments/k8s_run_e2e_test.bash features/eibanam/usermanagement/create_course.feature
# ./deployments/k8s_run_e2e_test.bash eibanam/usermanagement/create_course.feature
# ./deployments/k8s_run_e2e_test.bash usermanagement/create_course.feature

# Usage for testing one domain
# ./deployments/k8s_run_e2e_test.bash features/eibanam/lesson
# ./deployments/k8s_run_e2e_test.bash eibanam/lesson
# ./deployments/k8s_run_e2e_test.bash lesson

# Test all
# ./deployments/k8s_run_e2e_test.bash

SVCS=("communication" "lesson" "multitenant" "syllabus" "usermanagement" "entryexitmanagement")

# Runs e2e tests for the given feature file.
path="."
if [[ ! -z $1 ]]; then
  path=$1

  # Make "features/eibanam/lesson/some_file.feature" becomes "eibanam/lesson/some_file.feature"
  if [[ $path == features/* ]]; then
    path=${path:9}
  fi

  if [[ $path == eibanam/* ]]; then
    path=${path:8}
  fi

  svc=$(echo $path | cut -d '/' -f 1)
  if [[ ! " ${SVCS[@]} " =~ " ${svc} " ]]; then
    echo "Feature file is located in an invalid service \"$svc\", must be one of: ${SVCS[@]}"
    exit 1
  fi
  SVCS=($svc)
fi

# If run on local, enable tty so that we can send signal from terminal to the process inside the container.
# Cannot be done on Github Action (it's not a terminal).
export CI=${CI:-false}
tty_arg="--tty"
if [[ "$CI" == "true" ]]; then
  tty_arg=""
  echo "Running on CI, tty mode is disabled"
fi



svcForConfig="gandalf"
test_state=""
for svc in "${SVCS[@]}"; do
  gandalfPodName=$(kubectl -n backend get pods --selector=app.kubernetes.io/role="ci" --no-headers -o custom-columns=":metadata.name" --field-selector=status.phase=Running)

  pathToTest=""
  if [[ "$path" != "." ]]; then
    pathToTest=eibanam/${path}
  else
    pathToTest=eibanam/${svc}
  fi

  if [ "$gandalfPodName" == "" ]; then
    echo "No running pod found for gandalf to run test for eibanam"
    exit 1
  fi

  echo "=================== START TESTING $svc with $gandalfPodName"
  kubectl -n backend exec --stdin $tty_arg $gandalfPodName -c gandalf-ci -- sh -c "cd /backend/features/ && ./bdd.test \
    --godog.format=progress \
    --godog.tags=~@wip \
    --godog.concurrency=10 \
    -manabie.service=eibanam.$svc \
    -manabie.commonConfigPath=/configs/$svcForConfig.common.config.yaml \
    -manabie.configPath=/configs/$svcForConfig.config.yaml \
    -manabie.secretsPath=/configs/$svcForConfig.secrets.encrypted.yaml \
    -- $pathToTest
  "
  if [[ $? > 0 ]]; then
    test_state=failed
  fi
  echo "=================== END TESTING $svc"
done

if [[ "$test_state" == failed ]]; then
  echo "ERROR: some e2e tests failed"
  exit 1
fi
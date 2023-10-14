#!/bin/bash
set -e

export runnerNamespace="actions-runner-system"

case $ACTIONS in

  update)
    if kubectl -n $runnerNamespace get runnerdeployments | grep ${RUNNER_DEPLOY_NAME}; then
        echo "Update runner deployment: ${RUNNER_DEPLOY_NAME}";
        kubectl -n $runnerNamespace apply -f "./deployments/runner/${RUNNER_DEPLOY_NAME}.yaml"
    else
        echo "Cannot find runner deployment: ${RUNNER_DEPLOY_NAME}"
    fi
    ;;

  create)
    if ls -al "./deployments/runner/" | grep "${RUNNER_DEPLOY_NAME}.yaml"; then
        echo "Create runner deployment: ${RUNNER_DEPLOY_NAME}";
        kubectl -n $runnerNamespace create -f "./deployments/runner/${RUNNER_DEPLOY_NAME}.yaml"
    else
        echo "Cannot find runner deployment manifest file: ${RUNNER_DEPLOY_NAME}.yaml"
    fi
    ;;

  delete)
    if kubectl -n $runnerNamespace get runnerdeployments | grep ${RUNNER_DEPLOY_NAME}; then
        echo "Delete runner deployment: ${RUNNER_DEPLOY_NAME}";
        kubectl -n $runnerNamespace delete -f "./deployments/runner/${RUNNER_DEPLOY_NAME}.yaml"
    else
        echo "Cannot find runner deployment: ${RUNNER_DEPLOY_NAME}"
    fi
    ;;

  update_all)
    echo "Applying runner deployments"
    kubectl apply -f ./deployments/runner
    ;;

  *)
    echo -n "unknown action"
    ;;
esac
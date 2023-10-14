#!/bin/bash

# Usage:
#   deployments/helm/platforms/kafka/upgrade_kafka.bash

export ENV=${ENV:-local}
export ORG=${ORG:-manabie}
export LOCAL_REGISTRY_DOMAIN=${LOCAL_REGISTRY_DOMAIN:-localhost:5001}

DRY_RUN=${DRY_RUN:true}
namespace=${ENV}-${ORG}-kafka

# deployment_replicas is a map with [deployment]=#replicas
# example: deployment_replicas['kafka-connect']=1
declare -A deployment_replicas
disable_deployment_sequence=("cp-ksql-server" "kafka-connect" "cp-schema-registry")
# enable with reverse
enable_deployment_sequence=("cp-schema-registry" "kafka-connect" "cp-ksql-server")

function get_current_replicas() {
  local deployment="$1"
  local current_replicas=$(kubectl get deployment \
--namespace ${namespace} \
--no-headers \
--selector=app.kubernetes.io/instance="${deployment}" \
--output=custom-columns=":spec.replicas")
  return $current_replicas
}

function disable_kafka_clients() {
  for deployment in ${disable_deployment_sequence[@]}; do
    get_current_replicas $deployment
    replicas=$?
    deployment_replicas["${deployment}"]=${replicas}
    if [[ $replicas == 0 ]]; then
      echo "WARN: cannot get replicas value for ${deployment}"
    fi
    kubectl -n ${namespace} scale --replicas=0 deployment/${deployment}
  done
}

function enable_kafka_clients() {
  for deployment in ${enable_deployment_sequence[@]}; do
    replicas=${deployment_replicas[${deployment}]}
    if [[ $replicas == 0 ]]; then
      echo "WARN: cannot get replicas value for ${deployment}"
    fi
    kubectl -n ${namespace} scale --replicas=${replicas} deployment/${deployment}
    kubectl -n ${namespace} rollout status --timeout=3m deployment/${deployment}
  done
}

function upgrade_kafka() {
  skaffold deploy -f skaffold.backbone.yaml -p kafka-only
  # wait for kafka to completely deploy
  kubectl -n ${namespace} rollout status --timeout=3m statefulset/kafka
}

# main

./.github/scripts/diff_manifest.bash kafka
if [[ $DRY_RUN == "false" ]]; then
  disable_kafka_clients
  upgrade_kafka
  enable_kafka_clients
fi

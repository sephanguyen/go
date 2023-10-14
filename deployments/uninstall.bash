#!/bin/bash

set -eux

# Uninstall a helm release in a specific namespace
# Usage: helm_uninstall <namespace> <releaseName>
helm_uninstall() {
  namespace=$1
  release=$2
  if helm -n ${namespace} status ${release} >/dev/null 2>&1; then
    helm -n ${namespace} uninstall ${release}
  else
    echo "Skipping uninstalling ${release} in namespace ${namespace} (it may not exist)"
  fi
}

uninstall_gateway() {
  helm_uninstall istio-system ${ENV}-${ORG}-gateway
}

uninstall_services() {
  helm_uninstall ${ENV}-${ORG}-services manabie-all-in-one
}

uninstall_kakfa() {
  helm_uninstall ${ENV}-${ORG}-kafka kafka-connect
  helm_uninstall ${ENV}-${ORG}-kafka cp-schema-registry
  helm_uninstall ${ENV}-${ORG}-kafka kafka
  kubectl -n ${ENV}-${ORG}-kafka delete --ignore-not-found=true persistentvolumeclaims \
    kafka-pvc-kafka-0 \
    kafka-pvc-kafka-1 \
    kafka-pvc-kafka-2
}

uninstall_nats_jetstream() {
  helm_uninstall ${ENV}-${ORG}-nats-jetstream nats-jetstream
  kubectl -n ${ENV}-${ORG}-nats-jetstream delete --ignore-not-found=true persistentvolumeclaims \
    nats-jetstream-pvc-nats-jetstream-0 \
    nats-jetstream-pvc-nats-jetstream-1 \
    nats-jetstream-pvc-nats-jetstream-2
}

uninstall_elasticsearch() {
  releaseName="${ELASTIC_RELEASE_NAME:-elastic}"
  fullName=${ELASTIC_NAME_OVERRIDE:-${releaseName}}

  helm_uninstall ${ELASTIC_NAMESPACE} ${releaseName}
  kubectl -n ${ELASTIC_NAMESPACE} delete --ignore-not-found=true persistentvolumeclaims \
    elasticsearch-data-elasticsearch-${fullName}-0 \
    elasticsearch-data-elasticsearch-${fullName}-1 \
    elasticsearch-data-elasticsearch-${fullName}-2 \
    elasticsearch-snapshots-elasticsearch-${fullName}-0 \
    elasticsearch-snapshots-elasticsearch-${fullName}-1 \
    elasticsearch-snapshots-elasticsearch-${fullName}-2
}

uninstall_unleash() {
  helm uninstall "${ENV}-${ORG}-unleash" uninstall unleash
}

if [[ "$ENV" != "dorp" ]]; then
  >&2 echo "Script is reserved for dorp namespace only (current: ${ENV})"
  exit 1
fi

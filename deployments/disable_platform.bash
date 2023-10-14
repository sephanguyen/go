#!/bin/bash

set -eu

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

disable_nats_jetstream() {
  helm_uninstall ${ENV}-${ORG}-nats-jetstream nats-jetstream
}

disable_kakfa() {
  helm_uninstall ${ENV}-${ORG}-kafka kafka-connect
  helm_uninstall ${ENV}-${ORG}-kafka cp-schema-registry
  helm_uninstall ${ENV}-${ORG}-kafka kafka
}

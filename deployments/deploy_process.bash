#!/bin/bash

# PROCESS
#   M: That step must run manually
#   S: That step can be run by this script
#
#   1. Backup DB (M)
#   2. Run terraform apply to latest infras (M)
#     2.1 Config output service accounts to samena
#     2.2 Merge generated PR by samena to feature/synersia-prod-deployment (created from tag v0.32.0) to get latest secrets config.
#   3. Update new brighcove config (if exists) on feature/synersia-prod-deployment (M)
#   4. Tag v0.32.1 on feature/synersia-prod-deployment (M)
#   5. Install istio & cert-manager (S)
#   6. Deploy backend services using v.32.1 tag using Github Actions (M)
#   7. Install new istio gateway (S)
#   8. Cleanup old resources (gateways, virtual services) (S)
#   9. Install hasura for fatima (M)
#   10. Manually update virtual service for hasura fatima (M)

export ENV=prod
export ORG=${ORG:-synersia}
export NAMESPACE=${NAMESPACE:-backend}
export ISTIO_VERSION=1.8.2
ISTIO_VERSION_NAME=${ISTIO_VERSION//./-}

. ./deployments/setup_istio.bash

step5() {
  # delete old cert-manager version
  helm delete cert-manager -n cert-manager

  install_istio
  install_cert_manager

  # set istio rev point to new version
  kubectl label namespace $NAMESPACE istio.io/rev-
  kubectl label namespace $NAMESPACE istio.io/rev=${ISTIO_VERSION_NAME}
}

step7() {
  install_istio_gateway
}

step8() {
  helm delete gateways -n $NAMESPACE
}

case "$1" in
  step5)
    step5
    ;;
  step7)
    step7
    ;;
  step8)
    step8
    ;;
esac

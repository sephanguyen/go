#!/bin/bash

# This script verifies the current kubecontext matches with $ENV/$ORG
# by checking the existence of the expected namespaces.
#
# Usage:
#   ENV=stag ORG=manabie .github/tests/k8s_cluster.test.bash
#   ENV=prod ORG=jprep .github/tests/k8s_cluster.test.bash

set -euo pipefail

requiredNamespaces=(
    "istio-system"
    "cert-manager"
)
case $ENV in
    stag)
        case $ORG in
            manabie|jprep)
                requiredNamespaces+=(
                    "stag-manabie-services"
                    "stag-jprep-services"
                )
                ;;
            *)
                >&2 echo "[ERROR] Invalid <org> value \"$ORG\" for $ENV, must be one of: manabie, jprep"
                exit 1
                ;;
        esac
        ;;
    uat)
        case $ORG in
            manabie|jprep)
                requiredNamespaces+=(
                    "uat-manabie-services"
                    "uat-jprep-services"
                )
                ;;
            *)
                >&2 echo "[ERROR] Invalid <org> value \"$ORG\" for $ENV, must be one of: manabie, jprep"
                exit 1
                ;;
        esac
        ;;
    dorp|prod)
        case $ORG in
            manabie)
                ;;
            jprep)
                requiredNamespaces+=(
                    "prod-jprep-services"
                    "dorp-jprep-services"
                )
                ;;
            synersia|renseikai|ga|aic)
                requiredNamespaces+=(
                    "dorp-synersia-services"
                    "dorp-renseikai-services"
                    "dorp-ga-services"
                    "dorp-aic-services"
                    "prod-synersia-services"
                    "prod-renseikai-services"
                    "prod-ga-services"
                    "prod-aic-services"
                )
                ;;
            tokyo)
                requiredNamespaces+=(
                    # "dorp-tokyo-services"
                    "prod-tokyo-services"
                )
                ;;
            *)
                >&2 echo "[ERROR] Invalid <org> value \"$ORG\" for $ENV, must be one of: manabie, jprep, synersia, renseikai, ga, aic, tokyo"
                exit 1
                ;;
        esac
        ;;
    *)
        >&2 echo "[ERROR] Invalid <env> value \"$ENV\", must be one of: stag, uat, dorp, prod"
        exit 1
        ;;
esac

echo "Getting all namespaces"
kubectl get namespaces

for ns in "${requiredNamespaces[@]}"; do
    echo "Checking namespace $ns"
    if ! kubectl get namespace "$ns" >/dev/null; then
        >&2 echo "[ERROR] namespace $ns not found"
        exit 1
    fi
done

echo "Success"

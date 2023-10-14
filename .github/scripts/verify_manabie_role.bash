#!/bin/bash

# This script verify MANABIE role
# Requires package https://github.com/manabie-com/mana-packs/packages/948198.

set -euxo pipefail

hasuraServices=("bob" "eureka" "fatima" "timesheet" "entryexitmgmt" "invoicemgmt")

for service in "${hasuraServices[@]}"; do
    echo "Running for $service"
    pathToDB="$(pwd)/deployments/helm/manabie-all-in-one/charts/${service}/files/hasura/metadata/tables.yaml"
    yarn mana hasura manabie-role check \
        --metadata-path "${pathToDB}"
done

#!/bin/bash

# This script enables RLS for Hasura.
# Requires package https://github.com/manabie-com/mana-packs/packages/948198.

set -euxo pipefail

hasuraServices=($(python3 deployments/services-directory/hcl2hasura.py))

for service in "${hasuraServices[@]}"; do
    echo "Running for $service"
    pathToDB="$(pwd)/deployments/helm/manabie-all-in-one/charts/${service}/files/hasura/metadata"
    pathToIgnoreTables="$(pwd)/migrations/public_tables.json"
    if [ $service != "zeus" ] && [ $service != "draft" ] 
    then 
        yarn mana hasura permission check \
            --metadata-path "${pathToDB}"\
            --databases "${service}"\
            --ignore-table-path "${pathToIgnoreTables}"
    fi
done

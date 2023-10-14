#!/bin/bash

# This script enables RLS for Hasura.
# Requires package https://github.com/manabie-com/mana-packs/packages/948198.

set -euxo pipefail
npm set @manabie-com:registry=https://npm.pkg.github.com
npm set //npm.pkg.github.com/:_authToken="${GITHUB_TOKEN}"
npm install @manabie-com/mana-cli@0.6.2

#install python lib for parsing db name from hcl file
pip3 install python-hcl2 typing_extensions pyyaml

hasuraServices=($(python3 deployments/services-directory/hcl2hasura.py))

ignoreTables=(
    cities
    configs
    conversion_tasks
    districts
    dbz_signals
    debezium_signal
    lesson_schedules
    organizations
    organization_auths
    preset_study_plans_weekly_format
    schema_migration
    schools
    student_event_logs
    teacher_by_school_id
    prefecture
    granted_permissions
    configuration_group
    configuration_group_map
)

for service in "${hasuraServices[@]}"; do
    echo "Running for $service"
    pathToDB="$(pwd)/deployments/helm/manabie-all-in-one/charts/${service}/files/hasura/metadata"
    #ignore none user facing services
    if [ $service != "zeus" ] && [ $service != "draft" ] 
    then 
        yarn mana hasura permission rls \
            --ignoreTables "${ignoreTables[@]}" \
            --metadata-path "${pathToDB}"
    fi
done

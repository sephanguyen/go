#!/bin/bash

# This script update Timesheet service setting value for Partner in configturation table.
# Example: ./scripts/ad-hoc/sql/timesheet/migrate-on-off-timesheet-config-setting.bash

set -euo pipefail

DB_NAME="mastermgmt"


resource_path=""
config_value=""


print_usage() {
  printf "Usage: Update Timesheet service setting value for Partner in configturation table
    example: ./scripts/ad-hoc/sql/timesheet/migrate-on-off-timesheet-config-setting.bash -r -2147483648 -v on

    -r: resource_path or Partner
    -d: timesheet service config value for Partner (on/off)
  "
  
}

while getopts 'r:v:' flag; do
  case "${flag}" in
    r) resource_path="${OPTARG}" ;;
    v) config_value="${OPTARG}" ;;
    *) print_usage
       exit 1 ;;
  esac
done



psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
    UPDATE public."internal_configuration_value" set config_value = '${config_value}' where config_key = 'hcm.timesheet_management' and 
    resource_path = '${resource_path}'
    ; 
EOF

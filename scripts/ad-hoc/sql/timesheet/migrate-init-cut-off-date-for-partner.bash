#!/bin/bash

# This script use to init cut off date in Timesheet confirmation window for Partner
# Example: ./scripts/ad-hoc/sql/timesheet/migrate-init-cut-off-date-for-partner.bash

set -euo pipefail

DB_NAME="timesheet"

resource_path=""

print_usage() {
  printf "Usage: Init cut-off date for Partner
    example: ./scripts/ad-hoc/sql/timesheet/migrate-init-cut-off-date-for-partner.bash -r -2147483648 -d 12

    -r: resource_path or Partner
    -d: cut-off date of Partner (Ask PM or Partner for that)
  "
}

while getopts 'r:d:' flag; do
  case "${flag}" in
    r) resource_path="${OPTARG}" ;;
    d) new_cut_off_date="${OPTARG}" ;;
    *) print_usage
       exit 1 ;;
  esac
done

# insert to cut off date table
psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \ -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

    -- for run uuid_generate_v4 function
    CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

    INSERT INTO public."timesheet_confirmation_cut_off_date" 
    (id, cut_off_date, start_date, end_date, created_at, updated_at, resource_path) 
    VALUES 
    (uuid_generate_v4(), '$new_cut_off_date', '2021-12-31 15:00:00 +00:00' , '2099-12-30 14:59:59 +00:00', now(), now(), '$resource_path');
EOF

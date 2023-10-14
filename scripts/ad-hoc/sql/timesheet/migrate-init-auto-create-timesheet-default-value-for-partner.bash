#!/bin/bash

# This script use to init auto create timesheet default value for KEC Partner
# Example: ./scripts/ad-hoc/sql/timesheet/migrate-init-auto-create-timesheet-default-value-for-partner.bash

set -euo pipefail

DB_NAME="timesheet"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \ -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
    INSERT INTO public."partner_auto_create_timesheet_flag" 
    (id, flag_on, created_at, updated_at, resource_path) 
    VALUES 
    ('01GVEZNPFZ4GR51X044DSRXKG8', true, now(), now(), '-2147483642');
EOF

#!/bin/bash

# This script snapshot new captured table:
# - class
# - student_entrollment_status_history 

set -euo pipefail

DB_NAME="invoicemgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
    INSERT INTO dbz_signals(id, type, data) 
    VALUES ('snapshot-table-first-time-20230302', 'execute-snapshot', '{"data-collections": ["public.bank","public.bank_account"]}')
EOF

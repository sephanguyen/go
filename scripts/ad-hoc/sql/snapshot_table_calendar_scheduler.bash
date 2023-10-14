#!/bin/bash

# This script snapshot new captured table:
# - scheduler

set -euo pipefail

DB_NAME="calendar"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
    INSERT INTO tokyo_calendar.public.dbz_signals(id, type, data) 
    VALUES ('signal-beb34e75b9f54fa714cc212147b7fd74',  'execute-snapshot', '{"data-collections": ["public.scheduler"]}')
EOF

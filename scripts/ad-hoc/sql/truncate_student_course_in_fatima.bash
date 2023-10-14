#!/bin/bash

# This script truncates data in payment-related tables.

set -euo pipefail

DB_NAME="fatima"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
TRUNCATE public.student_course                           CASCADE;
EOF

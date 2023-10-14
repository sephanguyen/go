#!/bin/bash

# This script truncates data in some invoicemgmt-related tables.

set -euo pipefail

DB_NAME="invoicemgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
TRUNCATE public.bill_item CASCADE;
TRUNCATE public.invoice   CASCADE;
TRUNCATE public.discount  CASCADE;
TRUNCATE public.order     CASCADE;
EOF

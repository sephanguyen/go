#!/bin/bash

# This script updates the status of past SCHEDULED invoice schedule to `INVOICE_SCHEDULE_CANCELLED`

set -euo pipefail

DB_NAME="invoicemgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
UPDATE public.invoice_schedule
SET status = 'INVOICE_SCHEDULE_CANCELLED'
WHERE invoice_date < now() AND status = 'INVOICE_SCHEDULE_SCHEDULED';
EOF

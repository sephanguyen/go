#!/bin/bash

# This script imports an invoice schedule to a certain resource path

set -euo pipefail

DB_NAME="invoicemgmt"

INVOICE_SCHEDULE_ID=$1
ORG_ID=$2


psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

UPDATE invoice_schedule 
SET invoice_date = now(), scheduled_date = now(), updated_at = now(), remarks = 'updated for adhoc stress test'
WHERE invoice_schedule_id = '${INVOICE_SCHEDULE_ID}' 
AND resource_path = '${ORG_ID}';

EOF

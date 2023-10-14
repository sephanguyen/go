#!/bin/bash

# This script enables the payment sequence number trigger function

set -euo pipefail

DB_USER=$SA
DB_NAME="invoicemgmt"
DB_HOST="localhost"
DB_PORT="5432"

TOGGLE=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

ALTER TABLE payment ${TOGGLE} TRIGGER fill_in_payment_seq;

EOF
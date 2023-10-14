#!/bin/bash

# This script inserts values directly in company detail table

set -euo pipefail

DB_NAME="invoicemgmt"

COMPANY_DETAIL_ID=$1
COMPANY_NAME=$2
COMPANY_ADDRESS=$3
COMPANY_PHONE_NUMBER=$4
ORG_ID=$5

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

INSERT INTO public.company_detail(
    company_detail_id,
    company_name,
    company_address,
    company_phone_number,
    company_logo_url,
    created_at,
    updated_at,
    resource_path
) VALUES (
    '${COMPANY_DETAIL_ID}',
    '${COMPANY_NAME}',
    '${COMPANY_ADDRESS}',
    '${COMPANY_PHONE_NUMBER}',
    '',
    now(),
    now(),
    '${ORG_ID}'
);

EOF

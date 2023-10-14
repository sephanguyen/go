#!/bin/bash

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
UPDATE public.users
SET country = 'COUNTRY_JP'
WHERE
    resource_path = '-2147483644'
    AND user_group = 'USER_GROUP_SCHOOL_ADMIN'
    AND email ILIKE '%thu.vo+e2eschool%';
EOF

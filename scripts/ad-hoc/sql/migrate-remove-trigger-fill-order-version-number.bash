#!/bin/bash

set -euo pipefail

DB_NAME="fatima"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
DROP TRIGGER fill_in_order_version ON public.order;
EOF

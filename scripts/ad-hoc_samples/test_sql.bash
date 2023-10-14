#!/bin/bash

set -euo pipefail

DB_USER=$SA
DB_NAME="zeus"
DB_HOST="localhost"
DB_PORT="5432"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
\dt
SELECT * FROM $1 LIMIT $2;
EOF

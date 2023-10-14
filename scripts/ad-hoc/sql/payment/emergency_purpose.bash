#!/bin/bash

# This script hard delete permissions by ID

set -euo pipefail

DB_NAME="fatima"

STATEMENT=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

${STATEMENT}

EOF

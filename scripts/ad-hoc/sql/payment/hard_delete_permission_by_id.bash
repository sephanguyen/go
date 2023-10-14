#!/bin/bash

# This script hard delete permissions by ID

set -euo pipefail

DB_NAME="fatima"

PERMISSION_ID=$1
ORG_ID=$2

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

--- hard delete granted permission with permission ID ---
DELETE FROM granted_permission
WHERE permission_id = '${PERMISSION_ID}'
AND resource_path = '${ORG_ID}';

--- hard delete permission role ---
DELETE FROM permission_role
WHERE permission_id = '${PERMISSION_ID}'
AND resource_path = '${ORG_ID}';

--- hard delete permission ---
DELETE FROM permission
WHERE permission_id = '${PERMISSION_ID}'
AND resource_path = '${ORG_ID}';

EOF

#!/bin/bash

set -euo pipefail

DB_NAME="mastermgmt"

CONFIG_VALUE=$1
CONFIG_KEY=$2
ORG_ID=$3

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

UPDATE internal_configuration_value 
SET config_value = '${CONFIG_VALUE}', updated_at = now()
WHERE config_key = ANY('{${CONFIG_KEY}}')
AND resource_path = ANY('{${ORG_ID}}');

EOF

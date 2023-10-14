#!/bin/bash

# This script updates the payment order discount automation config value in a certain resource_path

set -euo pipefail

DB_NAME="mastermgmt"

CONFIG_VALUE=$1
ORG_ID=$2

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

UPDATE internal_configuration_value 
SET config_value = '${CONFIG_VALUE}', updated_at = now()
WHERE config_key = 'payment.order.enable_discount_automation' 
AND resource_path = '${ORG_ID}';

EOF

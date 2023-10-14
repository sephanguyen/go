#!/bin/bash

set -euo pipefail

DB_NAME="mastermgmt"

RESOURCE_PATH=$1
STATUS=$2

if ! [[ "$STATUS" == "on" || "$STATUS" == "off" ]]; then
  echo "STATUS is invalid. It should be 'on' or 'off'."
  exit 1
fi

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

UPDATE external_configuration_value
SET config_value = '${STATUS}'
WHERE 
       config_key = 'user.authentication.ip_address_restriction'
  AND resource_path = '${RESOURCE_PATH}';


-- show the result after update
SELECT config_key, config_value
FROM external_configuration_value
WHERE
       config_key = 'user.authentication.ip_address_restriction'
  AND resource_path = '${RESOURCE_PATH}';

EOF

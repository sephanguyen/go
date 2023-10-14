#!/bin/bash

set -euo pipefail

DB_NAME="mastermgmt"

# $1 = '-2147483648,-2147483647,-2147483645,-2147483640'
RESOURCE_PATHS=$1
# $2= '{"ipv4": ["127.0.0.1", "127.0.0.2"], "ipv6": ["2001:0db8:85a3:0000:0000:8a2e:0370:7334"]}'
CONFIG_VALUE=$2

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

UPDATE external_configuration_value
-- cast to JSON and then to TEXT for ensuring the value is json format
SET config_value = ('${CONFIG_VALUE}'::JSON)::TEXT
WHERE
       config_key = 'user.authentication.allowed_ip_address'
  AND resource_path = ANY('{${RESOURCE_PATHS}}');


-- show the result after update
SELECT config_key, config_value
FROM external_configuration_value
WHERE
       config_key = 'user.authentication.allowed_ip_address'
  AND resource_path = ANY('{${RESOURCE_PATHS}}');

EOF

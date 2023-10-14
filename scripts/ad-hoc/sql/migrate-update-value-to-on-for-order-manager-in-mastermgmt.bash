#!/bin/bash

# This script migrate Timesheet service setting value for Partner in configuration table.

set -euo pipefail

DB_MASTERMGMT_NAME="mastermgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_MASTERMGMT_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
UPDATE internal_configuration_value SET config_value = 'on'
WHERE resource_path = '-2147483639' AND config_key = 'payment.order.enable_order_manager';
EOF

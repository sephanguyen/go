#!/bin/bash

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

update class_member set start_date = now(), end_date = now() where start_date = '0001-01-01 00:00:00.000 +0000' and resource_path = '-2147483647';

EOF

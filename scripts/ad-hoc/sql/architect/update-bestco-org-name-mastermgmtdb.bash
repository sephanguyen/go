#!/bin/bash

set -euo pipefail

DB_NAME="mastermgmt"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

update organizations set name = 'Bestco' where organization_id = '-2147483644';
update organizations set name = '(Deprecated) Bestco' where organization_id = '-2147483643';

EOF

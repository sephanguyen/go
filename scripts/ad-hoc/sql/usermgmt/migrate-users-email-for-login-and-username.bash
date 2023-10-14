#!/bin/bash

set -euo pipefail

DB_NAME="bob"

RESOURCE_PATHS=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

-- migrate existing users to use the same email for login email
UPDATE users
SET login_email = email
WHERE
  login_email IS NULL AND
  resource_path = ANY('{${RESOURCE_PATHS}}');

-- migrate existing users to use the same email for username
UPDATE users
SET username = email
WHERE
  username IS NULL AND
  resource_path = ANY('{${RESOURCE_PATHS}}');

EOF

#!/bin/bash

set -euo pipefail

DB_NAME="bob"

ORG_ID=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
UPDATE users AS u
SET name = CONCAT(last_name,' ',first_name)
FROM students AS s
WHERE u.last_name != '' AND u.first_name != ''
AND CONCAT(last_name,' ',first_name) != name
AND u.deleted_at IS NULL
AND u.user_id = s.student_id
AND u.resource_path = ANY('{${ORG_ID}}');
EOF

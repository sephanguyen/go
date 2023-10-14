#!/bin/bash

set -euo pipefail

DB_NAME="bob"

ORG_ID=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
UPDATE users AS u
SET last_name = SPLIT_PART(last_name,CONCAT(' ',first_name),1)
FROM students AS s
WHERE last_name like CONCAT('% ',first_name)
AND CONCAT(last_name,' ',first_name) != previous_name
AND previous_name IS NOT NULL
AND u.user_id = s.student_id
AND u.deleted_at IS NULL
AND u.resource_path = ANY('{${ORG_ID}}');
EOF

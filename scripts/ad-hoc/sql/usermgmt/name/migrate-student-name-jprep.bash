#!/bin/bash

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
WITH user_first_names AS (SELECT user_id,CASE 
   WHEN first_name = '' and (given_name IS NOT NULL and given_name != '')
   THEN given_name
   ELSE first_name 
   END AS first_name
   FROM users)

UPDATE users AS u SET
previous_name = name,
last_name = name,
first_name = usrf.first_name,
name = CONCAT(name,' ',usrf.first_name)
FROM students AS s,user_first_names AS usrf
WHERE u.user_id = s.student_id
AND u.user_id = usrf.user_id
AND u.deleted_at IS NULL
AND u.resource_path = '-2147483647'; 
EOF

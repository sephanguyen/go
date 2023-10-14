#!/bin/bash

set -euo pipefail

DB_NAME="bob"

ORG_ID=$1
USER_ID=$2
NEW_EXTERNAL_USER_ID=$3

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
WITH existed_user as (
  SELECT count(*) as count_existed_user
  FROM users 
  WHERE resource_path = '${ORG_ID}'
  AND user_external_id  = '${NEW_EXTERNAL_USER_ID}'
)
UPDATE users 
SET user_external_id = '${NEW_EXTERNAL_USER_ID}'
WHERE user_id = '${USER_ID}'
AND resource_path = '${ORG_ID}'
AND (select count_existed_user from existed_user) = 0;
EOF

#!/bin/bash

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

WITH wrong_permission_role_in_other_tenant as (
  SELECT pr.permission_id, pr.role_id
  FROM permission_role pr
  JOIN permission p on p.permission_id = pr.permission_id
  WHERE p.resource_path != pr.resource_path
)
UPDATE permission_role
SET deleted_at = now()
WHERE permission_id IN (select w.permission_id from wrong_permission_role_in_other_tenant w)
AND role_id IN (select w.role_id from wrong_permission_role_in_other_tenant w);
EOF

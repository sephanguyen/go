#!/bin/bash

# This script updates configuration values by table name, config_key, config_value and resource_path

set -euo pipefail

DB_NAME="bob"

ORG_ID=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

-- deactivate parents who all of their children are deactivated or parents don't have children
update users set deactivated_at = now()
where
    user_group = 'USER_GROUP_PARENT' 
    and resource_path = ANY('{${ORG_ID}}')
    and user_id not in (
        select sp.parent_id from student_parents sp
            inner join users u on sp.student_id = u.user_id
        where sp.deleted_at is null and u.deactivated_at is null and sp.resource_path = ANY('{${ORG_ID}}')
        group by sp.parent_id 
);

-- deactivate parent who all of their children relationship are soft deleted
update users set deactivated_at = now()
where
    user_group = 'USER_GROUP_PARENT'
    and resource_path = ANY('{${ORG_ID}}')
    and user_id not in (
        select sp.parent_id from student_parents sp
        where sp.deleted_at is null and sp.resource_path = ANY('{${ORG_ID}}')
        group by sp.parent_id 
)
EOF

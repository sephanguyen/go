#!/bin/bash

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

update users set is_system = true where email = 'schedule_job+payment@manabie.com';

update users set user_role = 'student' where user_group = 'USER_GROUP_STUDENT';

update users set user_role = 'parent' where user_group = 'USER_GROUP_PARENT';

update users set user_role = 'staff'
where user_id in (select u.user_id from users u inner join staff s on u.user_id = s.staff_id);

update users set user_role = 'system' where is_system = true;
EOF

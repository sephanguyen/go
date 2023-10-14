#!/bin/bash

set -euo pipefail

DB_NAME="bob"

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF

delete from permission_role
where permission_id in (select permission_id from "permission" where permission_name = ANY('{user.user.write, user.parent.write, user.usergroupmember.write}'))
and role_id in (select role_id from "role" where role_name = ANY('{Teacher}'));

delete from permission_role
where permission_id in (select permission_id from "permission" where permission_name = ANY('{user.user.write, user.parent.write, user.staff.read}'))
and role_id in (select role_id from "role" where role_name = ANY('{Teacher Lead}'));

delete from permission_role
where permission_id in (select permission_id from "permission" where permission_name = ANY('{user.user.write, timesheet.timesheet.read, timesheet.timesheet.write, user.usergroup.write, user.usergroupmember.write, lesson.reallocation.write, lesson.lessonmember.write}'))
and role_id in (select role_id from "role" where role_name = ANY('{Student, Parent}'));

delete from permission_role
where permission_id in (select permission_id from "permission" where permission_name = ANY('{user.student.write, lesson.lesson.write}'))
and role_id in (select role_id from "role" where role_name = ANY('{Parent}'));
EOF

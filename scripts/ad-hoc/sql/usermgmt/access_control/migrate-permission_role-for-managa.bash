set -euo pipefail

DB_NAME="bob"

ORG_ID=$1

psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_PREFIX}${DB_NAME}" -p "${DB_PORT}" \
    -v ON_ERROR_STOP=1 --single-transaction --echo-all <<EOF
with 
role as (
    select r.role_id, r.resource_path
    from "role" r
    where r.role_name = 'UsermgmtScheduleJob'
    and r.resource_path = ANY ('{-2147483629, -2147483630}')),
permission as (
	select p.permission_id, p.resource_path
    from "permission" p 
    where p.permission_name = ANY ('{
        master.course.read,
        master.location.read,
        user.parent.read,
        user.parent.write,
        user.staff.read,
        user.staff.write,
        user.student_enrollment_status_history.read,
        user.student.read,
        user.student.write,
        user.usergroupmember.write,
        user.usergroupmember.read,
        user.usergroup.read,
        user.usergroup.write,
        user.user.read,
        user.user.write
    }')
	and p.resource_path = ANY ('{-2147483629, -2147483630}'))

insert into permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
select permission.permission_id,role.role_id, now(), now(), role.resource_path
  from role, permission
  where role.resource_path = permission.resource_path
  on conflict on constraint permission_role__pk do nothing;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAAFC4B05V')
    ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAACGYE8YA')
    ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;
EOF

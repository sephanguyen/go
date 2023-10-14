with 
role as (
  select r.role_id, r.resource_path
  from "role" r
  where r.role_name = 'Teacher'),
permission as (
	select p.permission_id, p.resource_path
  from "permission" p 
  where p.permission_name in (
    'user.staff.read'))

insert into permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
  select permission.permission_id,role.role_id, now(), now(), role.resource_path
  from role, permission
  where role.resource_path = permission.resource_path
  on conflict on constraint permission_role__pk do nothing;
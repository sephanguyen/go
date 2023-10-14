INSERT INTO permission
  (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
  ('01GM2DS7WP4S4JY7Q68P78NVZ7', 'master.location.write', NOW(), NOW(), '-2147483629')
  ON CONFLICT DO NOTHING;

with 
role as (
  select r.role_id, r.resource_path
  from "role" r
  where r.role_name = ANY ('{School Admin}')
	and r.resource_path = ANY ('{-2147483629}')),
permission as (
	select p.permission_id, p.resource_path
  from "permission" p 
  where p.permission_name = ANY ('{
	master.location.write}')
	and p.resource_path = ANY ('{-2147483629}'))

insert into permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
select permission.permission_id,role.role_id, now(), now(), role.resource_path
  from role, permission
  where role.resource_path = permission.resource_path
  on conflict on constraint permission_role__pk do nothing;

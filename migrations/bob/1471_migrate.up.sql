-- managara base: -2147483630
delete from granted_permission where permission_id = '01GGVRH1J0B91MSN1AHPT6YMBB';

delete from permission_role where permission_id = '01GGVRH1J0B91MSN1AHPT6YMBB';

delete from permission where permission_id = '01GGVRH1J0B91MSN1AHPT6YMBB';

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAAFC4B05V')
    ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

-- managara high school: -2147483629
delete from granted_permission where permission_id = '01GGVXWRZTN0N5ENQEV399KQH1';

delete from permission_role where permission_id = '01GGVXWRZTN0N5ENQEV399KQH1';

delete from permission where permission_id = '01GGVXWRZTN0N5ENQEV399KQH1';

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GJZQ8C9TCT9H3VWAACGYE8YA')
    ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

-- update table granted_permission
with 
role as (
  select r.role_id, r.resource_path
  from "role" r
  where r.role_name = 'UsermgmtScheduleJob'),
permission as (
	select p.permission_id, p.resource_path
  from "permission" p 
  where p.permission_name in (
    'master.location.read',
    'user.student.read',
    'user.student.write',
    'user.parent.read',
    'user.parent.write',
    'user.staff.read',
    'user.staff.write',
    'user.usergroup.read',
    'user.usergroup.write',
    'user.usergroupmember.write',
    'user.user.read',
    'user.user.write'))

insert into permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
select permission.permission_id,role.role_id, now(), now(), role.resource_path
  from role, permission
  where role.resource_path = permission.resource_path
  on conflict on constraint permission_role__pk do nothing;

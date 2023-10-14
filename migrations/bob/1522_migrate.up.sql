delete from permission_role 
where permission_id in (select permission_id from "permission" where permission_name = ANY('{user.user.write, user.parent.write}'))
and role_id in (select role_id from "role" where role_name = ANY('{Teacher}'));

delete from permission_role 
where permission_id in (select permission_id from "permission" where permission_name = ANY('{user.user.write, user.parent.write, user.staff.read}'))
and role_id in (select role_id from "role" where role_name = ANY('{Teacher Lead}'));

delete from permission_role 
where permission_id in (select permission_id from "permission" where permission_name = ANY('{user.user.write}'))
and role_id in (select role_id from "role" where role_name = ANY('{Student, Parent}'));

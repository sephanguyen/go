create or replace
view granted_permissions as
select
		ugm.user_id as user_id,
	 p.permission_name as permission_name,
	 l1.location_id as location_id
from
		user_group_member ugm
join user_group ug on
		ugm.user_group_id = ug.user_group_id
join granted_role gr on
		ug.user_group_id = gr.user_group_id
join "role" r on
		gr.role_id = r.role_id
join permission_role pr on
		r.role_id = pr.role_id
join "permission" p on
		p.permission_id = pr.permission_id
join granted_role_access_path grap on
		gr.granted_role_id = grap.granted_role_id
join locations l on
	l.location_id = grap.location_id
join locations l1 on
	l1.access_path ~ l.access_path
	and l.resource_path = l1.resource_path
where
	ugm.deleted_at is null
	and ug.deleted_at is null
	and gr.deleted_at is null
	and r.deleted_at is null
	and pr.deleted_at is null
	and p.deleted_at is null
	and grap.deleted_at is null
	and l.deleted_at is null
	and l1.deleted_at is null
;
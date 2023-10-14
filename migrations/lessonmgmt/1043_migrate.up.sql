DROP POLICY IF EXISTS rls_staff on "staff";
CREATE POLICY rls_staff_insert_location ON "staff" AS PERMISSIVE FOR INSERT TO PUBLIC
with check (
	1 = 1
);
CREATE POLICY rls_staff_select_location ON "staff" AS PERMISSIVE FOR select TO PUBLIC
using (
true <= (
	select			
		true
	from
					granted_permissions p
	join user_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'user.staff.read'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = staff.staff_id
		and usp.deleted_at is null
	limit 1
	)
)
;
CREATE POLICY rls_staff_update_location ON "staff" AS PERMISSIVE FOR update TO PUBLIC
using (
true <= (
	select			
		true
	from
					granted_permissions p
	join user_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'user.staff.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = staff.staff_id
		and usp.deleted_at is null
	limit 1
	)
)with check (
true <= (
	select			
		true
	from
					granted_permissions p
	join user_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'user.staff.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = staff.staff_id
		and usp.deleted_at is null
	limit 1
	)
)
;
CREATE POLICY rls_staff_delete_location ON "staff" AS PERMISSIVE FOR delete TO PUBLIC
using (
true <= (
	select			
		true
	from
					granted_permissions p
	join user_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_id = (
			select
				p2.permission_id
			from
				"permission" p2
			where
				p2.permission_name = 'user.staff.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = staff.staff_id
		and usp.deleted_at is null
	limit 1
	)
)
;

DROP POLICY IF EXISTS rls_students on "students";
CREATE POLICY rls_students_insert_location ON "students" AS PERMISSIVE FOR INSERT TO PUBLIC
with check (
	1 = 1
);
CREATE POLICY rls_students_select_location ON "students" AS PERMISSIVE FOR select TO PUBLIC
using (
student_id in (
	select			
		usp."user_id"
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
				p2.permission_name = 'user.student.read'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)
;
CREATE POLICY rls_students_update_location ON "students" AS PERMISSIVE FOR update TO PUBLIC
using (
student_id in (
	select			
		usp."user_id"
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
				p2.permission_name = 'user.student.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)with check (
student_id in (
	select			
		usp."user_id"
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
				p2.permission_name = 'user.student.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)
;
CREATE POLICY rls_students_delete_location ON "students" AS PERMISSIVE FOR delete TO PUBLIC
using (
student_id in (
	select			
		usp."user_id"
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
				p2.permission_name = 'user.student.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)
;

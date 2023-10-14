DROP POLICY IF EXISTS rls_courses on "courses";
DROP POLICY IF EXISTS rls_courses_permission_v3 on "courses";
CREATE POLICY rls_courses_permission_v3 ON "courses" AS PERMISSIVE FOR ALL TO PUBLIC
using (
	true <= (
		select			
			true
		from
			granted_permissions p
		where
			p.user_id = current_setting('app.user_id')
			and p.permission_id = (
				select
					p2.permission_id
				from
					"permission" p2
				where
					p2.permission_name = 'master.course.read'
					and p2.resource_path = current_setting('permission.resource_path'))
		limit 1
		)
)
with check (
	true <= (
		select			
			true
		from
			granted_permissions p
		where
			p.user_id = current_setting('app.user_id')
			and p.permission_id = (
				select
					p2.permission_id
				from
					"permission" p2
				where
					p2.permission_name = 'master.course.write'
					and p2.resource_path = current_setting('permission.resource_path'))
		limit 1
		)
);
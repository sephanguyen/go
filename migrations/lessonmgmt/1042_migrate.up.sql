DROP POLICY IF EXISTS rls_lessons on "lessons";
CREATE POLICY rls_lessons_location ON "lessons" AS PERMISSIVE FOR ALL TO PUBLIC
using (
	center_id in (
		select			
			p.location_id
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
					p2.permission_name = 'lesson.lesson.read'
					and p2.resource_path = current_setting('permission.resource_path'))
		)
)
with check (
	center_id in (
		select			
			p.location_id
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
					p2.permission_name = 'lesson.lesson.write'
					and p2.resource_path = current_setting('permission.resource_path'))
		)
);
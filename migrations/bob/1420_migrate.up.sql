DROP POLICY IF EXISTS rls_user_access_paths on "user_access_paths";
DROP POLICY IF EXISTS rls_user_access_paths_location on "user_access_paths";

CREATE POLICY rls_user_access_paths_location ON "user_access_paths" AS PERMISSIVE FOR ALL TO PUBLIC
using (
	location_id in (
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
					p2.permission_name = 'user.user.read'
					and p2.resource_path = current_setting('permission.resource_path'))
		)
)
with check (
	location_id in (
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
					p2.permission_name = 'user.user.write'
					and p2.resource_path = current_setting('permission.resource_path'))
		)
);
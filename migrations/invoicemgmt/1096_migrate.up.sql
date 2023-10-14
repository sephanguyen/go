DROP POLICY IF EXISTS rls_users on "users";
DROP POLICY IF EXISTS rls_users_insert_location on "users";
DROP POLICY IF EXISTS rls_users_select_location on "users";
DROP POLICY IF EXISTS rls_users_update_location on "users";
DROP POLICY IF EXISTS rls_users_delete_location on "users";

CREATE POLICY rls_users_insert_location ON "users" AS PERMISSIVE FOR INSERT TO PUBLIC
with check (
	1 = 1
);
CREATE POLICY rls_users_select_location ON "users" AS PERMISSIVE FOR select TO PUBLIC
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
				p2.permission_name = 'user.user.read'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = users.user_id
		and usp.deleted_at is null
	limit 1
	)
)
;
CREATE POLICY rls_users_update_location ON "users" AS PERMISSIVE FOR update TO PUBLIC
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
				p2.permission_name = 'user.user.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = users.user_id
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
				p2.permission_name = 'user.user.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = users.user_id
		and usp.deleted_at is null
	limit 1
	)
)
;
CREATE POLICY rls_users_delete_location ON "users" AS PERMISSIVE FOR delete TO PUBLIC
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
				p2.permission_name = 'user.user.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = users.user_id
		and usp.deleted_at is null
	limit 1
	)
)
;

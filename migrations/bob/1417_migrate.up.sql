DROP POLICY IF EXISTS rls_parents on "parents";
DROP POLICY IF EXISTS rls_parents_insert_location on "parents";
DROP POLICY IF EXISTS rls_parents_select_location on "parents";
DROP POLICY IF EXISTS rls_parents_update_location on "parents";
DROP POLICY IF EXISTS rls_parents_delete_location on "parents";

CREATE POLICY rls_parents_insert_location ON "parents" AS PERMISSIVE FOR INSERT TO PUBLIC
with check (
	1 = 1
);
CREATE POLICY rls_parents_select_location ON "parents" AS PERMISSIVE FOR select TO PUBLIC
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
				p2.permission_name = 'user.parent.read'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = parents.parent_id
		and usp.deleted_at is null
	limit 1
	)
)
;
CREATE POLICY rls_parents_update_location ON "parents" AS PERMISSIVE FOR update TO PUBLIC
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
				p2.permission_name = 'user.parent.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = parents.parent_id
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
				p2.permission_name = 'user.parent.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = parents.parent_id
		and usp.deleted_at is null
	limit 1
	)
)
;
CREATE POLICY rls_parents_delete_location ON "parents" AS PERMISSIVE FOR delete TO PUBLIC
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
				p2.permission_name = 'user.parent.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = parents.parent_id
		and usp.deleted_at is null
	limit 1
	)
)
;

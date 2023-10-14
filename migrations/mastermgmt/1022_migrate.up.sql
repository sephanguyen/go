DROP POLICY IF EXISTS rls_ac_test_template_1 on ac_test_template_1;
DROP POLICY IF EXISTS rls_ac_test_template_1_location on ac_test_template_1;
CREATE POLICY rls_ac_test_template_1_location ON ac_test_template_1 AS PERMISSIVE FOR ALL TO PUBLIC
using (
true <= (
	select			
		true
	from
					granted_permissions p
	join ac_test_template_1_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_name = 'accesscontrol.b.read'
		and usp."ac_test_template_1_id" = ac_test_template_1.ac_test_template_1_id
		and usp.deleted_at is null
	limit 1
	)
)
with check (
true <= (
	select			
		true
	from
					granted_permissions p
	join ac_test_template_1_access_paths usp on
					usp.location_id = p.location_id
	where
		p.user_id = current_setting('app.user_id')
		and p.permission_name = 'accesscontrol.b.write'
		and usp."ac_test_template_1_id" = ac_test_template_1.ac_test_template_1_id
		and usp.deleted_at is null
	limit 1
	)
);
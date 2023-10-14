DROP POLICY IF EXISTS rls_ac_test_template_4 on ac_test_template_4;
CREATE POLICY rls_ac_test_template_4_insert_permission_v4 ON ac_test_template_4 AS PERMISSIVE FOR INSERT TO PUBLIC
with check (
	1 = 1
);
CREATE POLICY rls_ac_test_template_4_select_permission_v4 ON ac_test_template_4 AS PERMISSIVE FOR select TO PUBLIC
using (
	current_setting('app.user_id') = owners
);
CREATE POLICY rls_ac_test_template_4_update_permission_v4 ON ac_test_template_4 AS PERMISSIVE FOR update TO PUBLIC
using (
	current_setting('app.user_id') = owners
)
with check (
	current_setting('app.user_id') = owners
);
CREATE POLICY rls_ac_test_template_4_delete_permission_v4 ON ac_test_template_4 AS PERMISSIVE FOR delete TO PUBLIC
using (
	current_setting('app.user_id') = owners
);

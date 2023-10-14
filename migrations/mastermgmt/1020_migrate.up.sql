DROP POLICY IF EXISTS rls_ac_test_template_11_4_insert_permission_v4 ON ac_test_template_11_4;
DROP POLICY IF EXISTS rls_ac_test_template_11_4_select_permission_v4 ON ac_test_template_11_4;
DROP POLICY IF EXISTS rls_ac_test_template_11_4_update_permission_v4 ON ac_test_template_11_4;
DROP POLICY IF EXISTS rls_ac_test_template_11_4_delete_permission_v4 ON ac_test_template_11_4;
CREATE POLICY rls_ac_test_template_11_4_permission_v4 ON ac_test_template_11_4 AS PERMISSIVE FOR ALL TO PUBLIC
using (
	current_setting('app.user_id') = owners
)
with check (
	current_setting('app.user_id') = owners
);
DROP POLICY IF EXISTS rls_users on "users";
DROP POLICY IF EXISTS rls_users_permission_v4 on "users";

CREATE POLICY rls_users_permission_v4 ON "users" AS PERMISSIVE FOR ALL TO PUBLIC
using (
	current_setting('app.user_id') = user_id
)
with check (
	current_setting('app.user_id') = user_id
);
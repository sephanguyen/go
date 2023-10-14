DROP POLICY IF EXISTS rls_timesheet on "timesheet";
CREATE POLICY rls_timesheet_permission_v4 ON "timesheet" AS PERMISSIVE FOR ALL TO PUBLIC
using (
	current_setting('app.user_id') = staff_id
)
with check (
	current_setting('app.user_id') = staff_id
);
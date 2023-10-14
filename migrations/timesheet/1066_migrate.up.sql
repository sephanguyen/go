DROP POLICY IF EXISTS rls_timesheet on "timesheet";
CREATE POLICY rls_timesheet_location ON "timesheet" AS PERMISSIVE FOR ALL TO PUBLIC
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
					p2.permission_name = 'timesheet.timesheet.read'
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
					p2.permission_name = 'timesheet.timesheet.write'
					and p2.resource_path = current_setting('permission.resource_path'))
		)
);
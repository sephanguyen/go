DROP POLICY IF EXISTS rls_payment on "payment";
CREATE POLICY rls_payment_location ON "payment" AS PERMISSIVE FOR ALL TO PUBLIC
using (
student_id in (
	select			
		usp."user_id"
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
				p2.permission_name = 'payment.payment.read'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
)
with check (
student_id in (
	select			
		usp."user_id"
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
				p2.permission_name = 'payment.payment.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp.deleted_at is null
	)
);
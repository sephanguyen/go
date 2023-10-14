DROP POLICY IF EXISTS rls_payment on "payment";
DROP POLICY IF EXISTS rls_payment_location on "payment";

CREATE POLICY rls_payment_location ON "payment" AS PERMISSIVE FOR ALL TO PUBLIC
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
				p2.permission_name = 'payment.payment.read'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = payment.student_id
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
		and usp."user_id" = payment.student_id
		and usp.deleted_at is null
	limit 1
	)
);
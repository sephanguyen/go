DROP POLICY IF EXISTS rls_billing_address on "billing_address";
DROP POLICY IF EXISTS rls_billing_address_location on "billing_address";

CREATE POLICY rls_billing_address_location ON "billing_address" AS PERMISSIVE FOR ALL TO PUBLIC
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
				p2.permission_name = 'payment.billing_address.read'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = billing_address.user_id
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
				p2.permission_name = 'payment.billing_address.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = billing_address.user_id
		and usp.deleted_at is null
	limit 1
	)
);
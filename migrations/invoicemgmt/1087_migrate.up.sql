DROP POLICY IF EXISTS rls_bank_account on "bank_account";
DROP POLICY IF EXISTS rls_bank_account_location on "bank_account";

CREATE POLICY rls_bank_account_location ON "bank_account" AS PERMISSIVE FOR ALL TO PUBLIC
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
				p2.permission_name = 'payment.bank_account.read'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = bank_account.student_id
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
				p2.permission_name = 'payment.bank_account.write'
				and p2.resource_path = current_setting('permission.resource_path'))
		and usp."user_id" = bank_account.student_id
		and usp.deleted_at is null
	limit 1
	)
);
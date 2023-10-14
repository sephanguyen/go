DROP POLICY IF EXISTS rls_student_product on "student_product";
CREATE POLICY rls_student_product_location ON "student_product" AS PERMISSIVE FOR ALL TO PUBLIC
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
					p2.permission_name = 'payment.student_product.read'
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
					p2.permission_name = 'payment.student_product.write'
					and p2.resource_path = current_setting('permission.resource_path'))
		)
);
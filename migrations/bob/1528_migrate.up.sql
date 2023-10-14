with 
role as (
  select r.role_id, r.resource_path
  from "role" r
  where r.role_name = ANY ('{OpenAPI}')
	and r.resource_path = ANY ('{-2147483642, -2147483644, 100000, -2147483627, -2147483640, -2147483630, -2147483628, -2147483646, -2147483643, -2147483635, -2147483631, -2147483626, -2147483648, -2147483647, -2147483634, -2147483645, -2147483637, -2147483641, -2147483638, -2147483629, -2147483639, -2147483625}')),
permission as (
	select p.permission_id, p.resource_path
  from "permission" p 
  where p.permission_name = ANY ('{
    payment.student_payment_detail.read,
    payment.student_payment_detail.write,
    payment.billing_address.read,
    payment.billing_address.write,
    payment.bank_account.read,
    payment.bank_account.write
	}')
	and p.resource_path = ANY ('{-2147483642, -2147483644, 100000, -2147483627, -2147483640, -2147483630, -2147483628, -2147483646, -2147483643, -2147483635, -2147483631, -2147483626, -2147483648, -2147483647, -2147483634, -2147483645, -2147483637, -2147483641, -2147483638, -2147483629, -2147483639, -2147483625}'))

insert into permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
select permission.permission_id,role.role_id, now(), now(), role.resource_path
  from role, permission
  where role.resource_path = permission.resource_path
  on conflict on constraint permission_role__pk do nothing;

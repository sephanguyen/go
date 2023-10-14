INSERT INTO permission
  (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
  ('01G1GQ13MVRCRHVP79GDPZTCY1', 'master.location.read', now(), now(), '-2147483639')
	ON CONFLICT DO NOTHING;

INSERT INTO role 
  (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
  ('01G1GQEKEHXKSM78NBW96NJ7K8', 'Teacher', false, now(), now(), '-2147483639'),
  ('01G1GQEKEHXKSM78NBW96NJ7K9', 'School Admin', false, now(), now(), '-2147483639'),
	('01G4EQH81VAK07Z3HN8WN87C19', 'Student', true, now(), now(), '-2147483639'),
  ('01G4EQH81VAK07Z3HN8WN87C20', 'Parent',  true, now(), now(), '-2147483639')
	ON CONFLICT DO NOTHING;

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES 
  ('01G1GQ13MVRCRHVP79GDPZTCY1', '01G1GQEKEHXKSM78NBW96NJ7K8', now(), now(), '-2147483639'),
  ('01G1GQ13MVRCRHVP79GDPZTCY1', '01G1GQEKEHXKSM78NBW96NJ7K9', now(), now(), '-2147483639'),
	('01G1GQ13MVRCRHVP79GDPZTCY1', '01G4EQH81VAK07Z3HN8WN87C19', now(), now(), '-2147483639'),
  ('01G1GQ13MVRCRHVP79GDPZTCY1', '01G4EQH81VAK07Z3HN8WN87C20', now(), now(), '-2147483639')
	ON CONFLICT DO NOTHING;

INSERT INTO public.user_group
  (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
  ('01G4EQH8SDX8QZ8FXC771Q6H19', 'Student', true, now(), now(), '-2147483639'),
  ('01G4EQH8SDX8QZ8FXC771Q6H20', 'Parent',  true, now(), now(), '-2147483639')

  ON CONFLICT DO NOTHING;

INSERT INTO public.granted_role
  (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01G4EQH8QNMSF8N93MQSNQNS19', '01G4EQH8SDX8QZ8FXC771Q6H19', '01G4EQH81VAK07Z3HN8WN87C19', now(), now(), '-2147483639'),
  ('01G4EQH8QNMSF8N93MQSNQNS20', '01G4EQH8SDX8QZ8FXC771Q6H20', '01G4EQH81VAK07Z3HN8WN87C20', now(), now(), '-2147483639')

  ON CONFLICT DO NOTHING;

INSERT INTO public.granted_role_access_path
  (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
  ('01G4EQH8QNMSF8N93MQSNQNS19', '01FR4M51XJY9E77GSN4QZ1Q8N1', now(), now(), '-2147483639'),
  ('01G4EQH8QNMSF8N93MQSNQNS20', '01FR4M51XJY9E77GSN4QZ1Q8N1', now(), now(), '-2147483639')

  ON CONFLICT DO NOTHING;

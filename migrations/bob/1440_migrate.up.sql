-- insert user_group
INSERT INTO public.user_group(
	user_group_id, user_group_name, created_at, updated_at, deleted_at, resource_path, org_location_id, is_system)
VALUES 
	('01G4EQH8SDX8QZ8FXC771Q1H01', 'UserGroup School Admin', now(), now(), null, '-2147483648', '01FR4M51XJY9E77GSN4QZ1Q9N1', true),
	('01G4EQH8SDX8QZ8FXC771Q1H02', 'UserGroup School Admin', now(), now(), null, '-2147483647', '01FR4M51XJY9E77GSN4QZ1Q9N2', true),
	('01G4EQH8SDX8QZ8FXC771Q1H05', 'UserGroup School Admin', now(), now(), null, '-2147483644', '01FR4M51XJY9E77GSN4QZ1Q9N5', true)
	ON CONFLICT DO NOTHING;
   
-- insert granted_role
INSERT INTO public.granted_role(
	granted_role_id, user_group_id, role_id, created_at, updated_at, deleted_at, resource_path)
VALUES 
	('01G4EQH8SDX8QZ8FXC791Q1H01', '01G4EQH8SDX8QZ8FXC771Q1H01', '01G1GQEKEHXKSM78NBW96NJ7H1', now(), now(), null, '-2147483648'),
	('01G4EQH8SDX8QZ8FXC791Q1H02', '01G4EQH8SDX8QZ8FXC771Q1H02', '01G1GQEKEHXKSM78NBW96NJ7H3', now(), now(), null, '-2147483647'),
	('01G4EQH8SDX8QZ8FXC791Q1H05', '01G4EQH8SDX8QZ8FXC771Q1H05', '01G1GQEKEHXKSM78NBW96NJ7H9', now(), now(), null, '-2147483644')
	ON CONFLICT DO NOTHING;

-- insert granted_role access_path
INSERT INTO public.granted_role_access_path(
	granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES
	('01G4EQH8SDX8QZ8FXC791Q1H01', '01FR4M51XJY9E77GSN4QZ1Q9N1', now(), now(), null, '-2147483648'),
	('01G4EQH8SDX8QZ8FXC791Q1H02', '01FR4M51XJY9E77GSN4QZ1Q9N2', now(), now(), null, '-2147483647'),
	('01G4EQH8SDX8QZ8FXC791Q1H05', '01FR4M51XJY9E77GSN4QZ1Q9N5', now(), now(), null, '-2147483644')
	ON CONFLICT DO NOTHING;

--- insert user
INSERT INTO public.users
(user_id, country, "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES
	('bdd_admin+manabie', 'COUNTRY_JP', 'bdd_admin+manabie@manabie.com', '', 'bdd_admin+manabie@manabie.com', 'bdd_admin+manabie@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483648', true),
	('bdd_admin+jprep', 'COUNTRY_JP', 'bdd_admin+jprep@manabie.com', '', 'bdd_admin+jprep@manabie.com', 'bdd_admin+jprep@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483647', true),
	('bdd_admin+e2e', 'COUNTRY_JP', 'bdd_admin+e2e@manabie.com', '', 'bdd_admin+e2e@manabie.com', 'bdd_admin+e2e@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483644', true)
	ON CONFLICT DO NOTHING;

-- user_group_member
INSERT INTO public.user_group_member(
	user_id, user_group_id, created_at, updated_at, deleted_at, resource_path)
VALUES 
	('bdd_admin+manabie', '01G4EQH8SDX8QZ8FXC771Q1H01', now(), now(), null, '-2147483648'),
	('bdd_admin+jprep', '01G4EQH8SDX8QZ8FXC771Q1H02', now(), now(), null, '-2147483647'),
	('bdd_admin+e2e', '01G4EQH8SDX8QZ8FXC771Q1H05', now(), now(), null, '-2147483644')
	ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths(
	user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES
	('bdd_admin+manabie', '01FR4M51XJY9E77GSN4QZ1Q9N1', null, now(), now(), null, '-2147483648'),
	('bdd_admin+jprep', '01FR4M51XJY9E77GSN4QZ1Q9N2', null, now(), now(), null, '-2147483647'),
	('bdd_admin+e2e', '01FR4M51XJY9E77GSN4QZ1Q9N5', null, now(), now(), null, '-2147483644')
	ON CONFLICT DO NOTHING;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01G4EQH8SDX8QZ8FXC771Q1H01')
    ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01G4EQH8SDX8QZ8FXC771Q1H02')
    ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01G4EQH8SDX8QZ8FXC771Q1H05')
    ON CONFLICT ON CONSTRAINT granted_permission__pk DO NOTHING;

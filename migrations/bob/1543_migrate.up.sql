--- Create permission for E2E-Architecture ---

INSERT INTO role 
  (role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES 
  ('01H04ZMKB06FD7950JX46ET5S7', 'UsermgmtScheduleJob',  true, now(), now(), '100000')
  ON CONFLICT DO NOTHING;

INSERT INTO permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('911FLMNMYA1J2ZWE640JQBMJ0J', '01H04ZMKB06FD7950JX46ET5S7', now(), now(), '100000'),
  ('01GMMSRDYPDX9H0XMVHKBCFD49', '01H04ZMKB06FD7950JX46ET5S7', now(), now(), '100000'),
  ('01GMMSRDYPDX9H0XMVHCKEHVQ4', '01H04ZMKB06FD7950JX46ET5S7', now(), now(), '100000'),
  ('01GMMSRDYPDX9H0XMVGC65BXH8', '01H04ZMKB06FD7950JX46ET5S7', now(), now(), '100000')
  ON CONFLICT DO NOTHING;

INSERT INTO public.user_group
  (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
  ('01H04ZRZCS6W8BMR3T0PF3E083', 'UsermgmtScheduleJob', true, now(), now(), '100000')
  ON CONFLICT DO NOTHING;

INSERT INTO public.granted_role
  (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01H04ZV5X98TAQG7CF80Q11GC1', '01H04ZRZCS6W8BMR3T0PF3E083', '01H04ZMKB06FD7950JX46ET5S7', now(), now(), '100000')
  ON CONFLICT DO NOTHING;

INSERT INTO public.granted_role_access_path
  (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
  ('01H04ZV5X98TAQG7CF80Q11GC1', '911FLMNMYA6SKTTT44HWE2E100', now(), now(), '100000')
  ON CONFLICT DO NOTHING;

-- Add Usermgmt user for Schedule Job --

-- resource_path 100000 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01H04ZYG5F9NGV4VNH865627H4', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '100000', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01H04ZYG5F9NGV4VNH865627H4', '01H04ZRZCS6W8BMR3T0PF3E083', now(), now(), '100000') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01H04ZYG5F9NGV4VNH865627H4', now(), now(), NULL, '100000', DEFAULT, DEFAULT, NULL, NULL) ON CONFLICT DO NOTHING;

-- resource_path -2147483623 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01H05014ZVWAQRG8V8GZ1F5MXR', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483623', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01H05014ZVWAQRG8V8GZ1F5MXR', '01GZZJS55NA1XFHWM1PF8HX130', now(), now(), '-2147483623') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01H05014ZVWAQRG8V8GZ1F5MXR', now(), now(), NULL, '-2147483623', DEFAULT, DEFAULT, NULL, NULL) ON CONFLICT DO NOTHING;

-- resource_path -2147483624 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01H050217ANRWWNX21J2D18AS2', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483624', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01H050217ANRWWNX21J2D18AS2', '01GZJEWEAWTT4FXYJMG1KQ9XV2', now(), now(), '-2147483624') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01H050217ANRWWNX21J2D18AS2', now(), now(), NULL, '-2147483624', DEFAULT, DEFAULT, NULL, NULL) ON CONFLICT DO NOTHING;

-- resource_path -2147483625 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01H05033TZ5G6J919JKYYV4ZJN', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483625', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01H05033TZ5G6J919JKYYV4ZJN', '01GY9K6FC9K7TG1FQ92XS7TFAG', now(), now(), '-2147483625') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01H05033TZ5G6J919JKYYV4ZJN', now(), now(), NULL, '-2147483625', DEFAULT, DEFAULT, NULL, NULL) ON CONFLICT DO NOTHING;

-- resource_path -2147483626 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01H0504BD93VC9QZY0GN0WPN3K', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483626', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01H0504BD93VC9QZY0GN0WPN3K', '01GX2YXGKDGVQT7GT8ANJMWS99', now(), now(), '-2147483626') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01H0504BD93VC9QZY0GN0WPN3K', now(), now(), NULL, '-2147483626', DEFAULT, DEFAULT, NULL, NULL) ON CONFLICT DO NOTHING;

-- resource_path -2147483627 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01H0505237SAJ9PZS07X8HE78C', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483627', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01H0505237SAJ9PZS07X8HE78C', '01GTBD78816KTXZJ2WC9Q2VGP2', now(), now(), '-2147483627') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01H0505237SAJ9PZS07X8HE78C', now(), now(), NULL, '-2147483627', DEFAULT, DEFAULT, NULL, NULL) ON CONFLICT DO NOTHING;

-- resource_path -2147483628 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01H0505SB9Y7FXK290CYFDJ86R', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483628', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01H0505SB9Y7FXK290CYFDJ86R', '01GRB9AE84VMPC803E3FNJXH82', now(), now(), '-2147483628') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01H0505SB9Y7FXK290CYFDJ86R', now(), now(), NULL, '-2147483628', DEFAULT, DEFAULT, NULL, NULL) ON CONFLICT DO NOTHING;

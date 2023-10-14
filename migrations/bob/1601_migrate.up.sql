--- Add permission ---
INSERT INTO permission
(permission_id, permission_name, created_at, updated_at, resource_path)
VALUES
    ('01GTZYYSBEEVBQXREAYRR34PA1', 'payment.product.read', now(), now(), '-2147483622'),
    ('01GTZYYSBEEVBQXREAYRR34PA2', 'payment.product.write', now(), now(), '-2147483622'),
    ('01GTZYYSBEEVBQXREAYRR34PB1', 'payment.product.read', now(), now(), '-2147483623'),
    ('01GTZYYSBEEVBQXREAYRR34PB2', 'payment.product.write', now(), now(), '-2147483623'),
    ('01GTZYYSBEEVBQXREAYRR34QA1', 'payment.product.read', now(), now(), '-2147483624'),
    ('01GTZYYSBEEVBQXREAYRR34QA2', 'payment.product.write', now(), now(), '-2147483624'),
    ('01GTZYYSBEEVBQXREAYRR34XB1', 'payment.product.read', now(), now(), '-2147483625'),
    ('01GTZYYSBEEVBQXREAYRR34XB2', 'payment.product.write', now(), now(), '-2147483625'),
    ('01GTZYYSBEEVBQXREAYRR34YA1', 'payment.product.read', now(), now(), '-2147483626'),
    ('01GTZYYSBEEVBQXREAYRR34YA2', 'payment.product.write', now(), now(), '-2147483626')
    ON CONFLICT DO NOTHING;

--- Add role ---
INSERT INTO role
(role_id, role_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01GTZYYSEPYTSDW16YPQYYSEP2', 'PaymentScheduleJob', true, now(), now(), '-2147483622'),
    ('01GTZYYSEPYTSDW16YPQYYSEP3', 'PaymentScheduleJob', true, now(), now(), '-2147483623'),
    ('01GTZYYSEPYTSDW16YPQYYSEP4', 'PaymentScheduleJob', true, now(), now(), '-2147483624'),
    ('01GTZYYSEPYTSDW16YPQYYSEP5', 'PaymentScheduleJob', true, now(), now(), '-2147483625'),
    ('01GTZYYSEPYTSDW16YPQYYSEP6', 'PaymentScheduleJob', true, now(), now(), '-2147483626')
    ON CONFLICT DO NOTHING;

--- Add permission_role ---
INSERT INTO permission_role
(permission_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01GTZYYSBEEVBQXREAYRR34PA1', '01GTZYYSEPYTSDW16YPQYYSEP2', now(), now(), '-2147483622'),
    ('01GTZYYSBEEVBQXREAYRR34PA2', '01GTZYYSEPYTSDW16YPQYYSEP2', now(), now(), '-2147483622'),
    ('01GTZYYSBEEVBQXREAYRR34PB1', '01GTZYYSEPYTSDW16YPQYYSEP3', now(), now(), '-2147483623'),
    ('01GTZYYSBEEVBQXREAYRR34PB2', '01GTZYYSEPYTSDW16YPQYYSEP3', now(), now(), '-2147483623'),
    ('01GTZYYSBEEVBQXREAYRR34QA1', '01GTZYYSEPYTSDW16YPQYYSEP4', now(), now(), '-2147483624'),
    ('01GTZYYSBEEVBQXREAYRR34QA2', '01GTZYYSEPYTSDW16YPQYYSEP4', now(), now(), '-2147483624'),
    ('01GTZYYSBEEVBQXREAYRR34XB1', '01GTZYYSEPYTSDW16YPQYYSEP5', now(), now(), '-2147483625'),
    ('01GTZYYSBEEVBQXREAYRR34XB2', '01GTZYYSEPYTSDW16YPQYYSEP5', now(), now(), '-2147483625'),
    ('01GTZYYSBEEVBQXREAYRR34YA1', '01GTZYYSEPYTSDW16YPQYYSEP6', now(), now(), '-2147483626'),
    ('01GTZYYSBEEVBQXREAYRR34YA2', '01GTZYYSEPYTSDW16YPQYYSEP6', now(), now(), '-2147483626')
    ON CONFLICT DO NOTHING;

--- Add User Group ---
INSERT INTO public.user_group
(user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
    ('01GV00DNYKP9CDSD30MMQF2V02', 'PaymentScheduleJob', true, now(), now(), '-2147483622'),
    ('01GV00DNYKP9CDSD30MMQF2V03', 'PaymentScheduleJob', true, now(), now(), '-2147483623'),
    ('01GV00DNYKP9CDSD30MMQF2V04', 'PaymentScheduleJob', true, now(), now(), '-2147483624'),
    ('01GV00DNYKP9CDSD30MMQF2V05', 'PaymentScheduleJob', true, now(), now(), '-2147483625'),
    ('01GV00DNYKP9CDSD30MMQF2V06', 'PaymentScheduleJob', true, now(), now(), '-2147483626')
    ON CONFLICT DO NOTHING;

--- Grant role to User group ---
INSERT INTO public.granted_role
(granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
    ('01GFWRE5WXWA24F8TVXJHTWRE2', '01GV00DNYKP9CDSD30MMQF2V02', '01GTZYYSEPYTSDW16YPQYYSEP2', now(), now(), '-2147483622'),
    ('01GFWRE5WXWA24F8TVXJHTWRE3', '01GV00DNYKP9CDSD30MMQF2V03', '01GTZYYSEPYTSDW16YPQYYSEP3', now(), now(), '-2147483623'),
    ('01GFWRE5WXWA24F8TVXJHTWRE4', '01GV00DNYKP9CDSD30MMQF2V04', '01GTZYYSEPYTSDW16YPQYYSEP4', now(), now(), '-2147483624'),
    ('01GFWRE5WXWA24F8TVXJHTWRE5', '01GV00DNYKP9CDSD30MMQF2V05', '01GTZYYSEPYTSDW16YPQYYSEP5', now(), now(), '-2147483625'),
    ('01GFWRE5WXWA24F8TVXJHTWRE6', '01GV00DNYKP9CDSD30MMQF2V06', '01GTZYYSEPYTSDW16YPQYYSEP6', now(), now(), '-2147483626')
    ON CONFLICT DO NOTHING;

--- Grant location to a role ---
INSERT INTO public.granted_role_access_path
(granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
    ('01GFWRE5WXWA24F8TVXJHTWRE2', '01H1BAHG1YXMMXEVSTS7NBV9CH', now(), now(), '-2147483622'),
    ('01GFWRE5WXWA24F8TVXJHTWRE3', '01GZZJE0A5AMPCCF41GA830NEK', now(), now(), '-2147483623'),
    ('01GFWRE5WXWA24F8TVXJHTWRE4', '01GZJDV9GCHMYD8MKGT7CV052W', now(), now(), '-2147483624'),
    ('01GFWRE5WXWA24F8TVXJHTWRE5', '01GY9KYKRW4M16YPB0V03JRTE6', now(), now(), '-2147483625'),
    ('01GFWRE5WXWA24F8TVXJHTWRE6', '01GX2R4MH7FXFKDH26JKMV91Q5', now(), now(), '-2147483626')
    ON CONFLICT DO NOTHING;

--- Upsert granted_permission ---
INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF2V02')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF2V03')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF2V04')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF2V05')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;

INSERT INTO granted_permission
(user_group_id, user_group_name, role_id, role_name, permission_id, permission_name, location_id, resource_path)
SELECT * FROM retrieve_src_granted_permission('01GV00DNYKP9CDSD30MMQF2V05')
    ON CONFLICT ON CONSTRAINT granted_permission__pk
    DO UPDATE SET user_group_name = excluded.user_group_name;


-- resource_path -2147483622 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path)
VALUES('01GTZYX224982Z1X4MHZQW6BO2', 'COUNTRY_JP', 'Payment Schedule Job', '', NULL, 'schedule_job+payment@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483622')
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GTZYX224982Z1X4MHZQW6BO2', '01GV00DNYKP9CDSD30MMQF2V02', now(), now(), '-2147483622') ON CONFLICT DO NOTHING;

-- resource_path -2147483623 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path)
VALUES('01GTZYX224982Z1X4MHZQW6BO3', 'COUNTRY_JP', 'Payment Schedule Job', '', NULL, 'schedule_job+payment@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483623')
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GTZYX224982Z1X4MHZQW6BO3', '01GV00DNYKP9CDSD30MMQF2V03', now(), now(), '-2147483623') ON CONFLICT DO NOTHING;

-- resource_path -2147483624 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path)
VALUES('01GTZYX224982Z1X4MHZQW6BO4', 'COUNTRY_JP', 'Payment Schedule Job', '', NULL, 'schedule_job+payment@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483624')
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GTZYX224982Z1X4MHZQW6BO4', '01GV00DNYKP9CDSD30MMQF2V04', now(), now(), '-2147483624') ON CONFLICT DO NOTHING;

-- resource_path -2147483625 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path)
VALUES('01GTZYX224982Z1X4MHZQW6BO5', 'COUNTRY_JP', 'Payment Schedule Job', '', NULL, 'schedule_job+payment@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483625')
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GTZYX224982Z1X4MHZQW6BO5', '01GV00DNYKP9CDSD30MMQF2V05', now(), now(), '-2147483625') ON CONFLICT DO NOTHING;

-- resource_path -2147483626 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path)
VALUES('01GTZYX224982Z1X4MHZQW6BO6', 'COUNTRY_JP', 'Payment Schedule Job', '', NULL, 'schedule_job+payment@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483626')
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GTZYX224982Z1X4MHZQW6BO6', '01GV00DNYKP9CDSD30MMQF2V06', now(), now(), '-2147483626') ON CONFLICT DO NOTHING;

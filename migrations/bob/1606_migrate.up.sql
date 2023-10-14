-- synersia-internal 2147483646--
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H4JC53MZ03T79DNR02YFQNHH', true, now(), now(), NULL, '2147483646') ON CONFLICT DO NOTHING;

-- eishinkan-internal 2147483631--
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H4QGKSC901ZJWG4C03XVKME5', true, now(), now(), NULL, '2147483631') ON CONFLICT DO NOTHING;

-- withus-base-internal 2147483630--
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H4Z1V2XB039NEVFB02VDDNBZ', true, now(), now(), NULL, '2147483630') ON CONFLICT DO NOTHING;

-- withus-hs-internal 2147483629--
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H4Z60EJY01A5J8XM02J6JRRS', true, now(), now(), NULL, '2147483629') ON CONFLICT DO NOTHING;

-- manabie demo erp -2147483628--
-- users
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01H5QEXMDZS7JZC3Q91S931NRW', 'COUNTRY_JP', 'Notification Schedule Job', '', NULL, 'schedule_job+notification@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483628', true)
    ON CONFLICT DO NOTHING;
-- user_group_member
INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01H5QEXMDZS7JZC3Q91S931NRW', '01GRB9AE84VMPC803E3FNJXH81', now(), now(), '-2147483628') ON CONFLICT DO NOTHING;
-- staff
INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01H5QEXMDZS7JZC3Q91S931NRW', now(), now(), NULL, '-2147483628', DEFAULT, DEFAULT, NULL, NULL) ON CONFLICT DO NOTHING;
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H5QEXMDZS7JZC3Q91S931NRW', true, now(), now(), NULL, '-2147483628') ON CONFLICT DO NOTHING;

-- kec lms pilot -2147483627--
-- users
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01H5QFHDM2N55ESJP7NYKPG1B5', 'COUNTRY_JP', 'Notification Schedule Job', '', NULL, 'schedule_job+notification@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483627', true)
    ON CONFLICT DO NOTHING;
-- user_group_member
INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01H5QFHDM2N55ESJP7NYKPG1B5', '01GTBD78816KTXZJ2WC9Q2VGP1', now(), now(), '-2147483627') ON CONFLICT DO NOTHING;
-- staff
INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01H5QFHDM2N55ESJP7NYKPG1B5', now(), now(), NULL, '-2147483627', DEFAULT, DEFAULT, NULL, NULL) ON CONFLICT DO NOTHING;
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H5QFHDM2N55ESJP7NYKPG1B5', true, now(), now(), NULL, '-2147483627') ON CONFLICT DO NOTHING;

-- keisin -2147483626--
-- users
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01H5QFV9ERVF0Y26JHF5ZTRHZA', 'COUNTRY_JP', 'Notification Schedule Job', '', NULL, 'schedule_job+notification@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483626', true)
    ON CONFLICT DO NOTHING;
-- user_group_member
INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01H5QFV9ERVF0Y26JHF5ZTRHZA', '01GX2YXA48N2QNDMWB5YX5KSQP', now(), now(), '-2147483626') ON CONFLICT DO NOTHING;
-- staff
INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01H5QFV9ERVF0Y26JHF5ZTRHZA', now(), now(), NULL, '-2147483626', DEFAULT, DEFAULT, NULL, NULL) ON CONFLICT DO NOTHING;
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H5QFV9ERVF0Y26JHF5ZTRHZA', true, now(), now(), NULL, '-2147483626') ON CONFLICT DO NOTHING;

-- keisin-internal 2147483626--
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H4Z64D5M00YK0GPC01DN9P1Z', true, now(), now(), NULL, '2147483626') ON CONFLICT DO NOTHING;

-- seiki -2147483625--
-- users
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01H5QGSS7G266KPJH42K71T7SE', 'COUNTRY_JP', 'Notification Schedule Job', '', NULL, 'schedule_job+notification@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483625', true)
    ON CONFLICT DO NOTHING;
-- user_group_member
INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01H5QGSS7G266KPJH42K71T7SE', '01GY9K6ANSW1290ZPAMN48YQZZ', now(), now(), '-2147483625') ON CONFLICT DO NOTHING;
-- staff
INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01H5QGSS7G266KPJH42K71T7SE', now(), now(), NULL, '-2147483625', DEFAULT, DEFAULT, NULL, NULL) ON CONFLICT DO NOTHING;
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H5QGSS7G266KPJH42K71T7SE', true, now(), now(), NULL, '-2147483625') ON CONFLICT DO NOTHING;

-- seiki-internal 2147483625--
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H4Z66YAF008QR7BZ00JCXZ9X', true, now(), now(), NULL, '2147483625') ON CONFLICT DO NOTHING;

-- withus-juku-internal 2147483624--
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H4Z91ZB600WDFTBR01EWVKNP', true, now(), now(), NULL, '2147483624') ON CONFLICT DO NOTHING;

-- kec testing -2147483623--
-- users
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01H5QHVFPD8P6V4NEJ0CR0JGP1', 'COUNTRY_JP', 'Notification Schedule Job', '', NULL, 'schedule_job+notification@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483623', true)
    ON CONFLICT DO NOTHING;
-- user_group_member
INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01H5QHVFPD8P6V4NEJ0CR0JGP1', '01GZZJS090YNZJSW995F4ZZJ33', now(), now(), '-2147483623') ON CONFLICT DO NOTHING;
-- staff
INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01H5QHVFPD8P6V4NEJ0CR0JGP1', now(), now(), NULL, '-2147483623', DEFAULT, DEFAULT, NULL, NULL) ON CONFLICT DO NOTHING;
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H5QHVFPD8P6V4NEJ0CR0JGP1', true, now(), now(), NULL, '-2147483623') ON CONFLICT DO NOTHING;

-- lms v2 -2147483622--
-- users
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01H5QJ6GRDQKY667T05A8G84DB', 'COUNTRY_JP', 'Notification Schedule Job', '', NULL, 'schedule_job+notification@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483622', true)
    ON CONFLICT DO NOTHING;
-- user_group_member
INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01H5QJ6GRDQKY667T05A8G84DB', '01H1BAXNXV4KFKCC6KS8EQJ0EP', now(), now(), '-2147483622') ON CONFLICT DO NOTHING;
-- staff
INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01H5QJ6GRDQKY667T05A8G84DB', now(), now(), NULL, '-2147483622', DEFAULT, DEFAULT, NULL, NULL) ON CONFLICT DO NOTHING;
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H5QJ6GRDQKY667T05A8G84DB', true, now(), now(), NULL, '-2147483622') ON CONFLICT DO NOTHING;

-- renseikai-internal 2147483645--
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H4QBARW3007HV2YD01AQENVG', true, now(), now(), NULL, '2147483645') ON CONFLICT DO NOTHING;

-- bestco-internal 2147483643--
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H4QG7EKS01TPVSK0003CB7ER', true, now(), now(), NULL, '2147483643') ON CONFLICT DO NOTHING;

-- aic-internal 2147483641--
-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H4QBRYT000YNBXY502MKN06A', true, now(), now(), NULL, '2147483641') ON CONFLICT DO NOTHING;
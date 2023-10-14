-- resource_path -2147483624 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01H3K88JTY9P7QQBQ16Z63NZDK', 'COUNTRY_JP', 'Notification Schedule Job', '', NULL, 'schedule_job+notification@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483624', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01H3K88JTY9P7QQBQ16Z63NZDK', '01GZJEWM58G9S3GDH4X3T03YV2', now(), now(), '-2147483624') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01H3K88JTY9P7QQBQ16Z63NZDK', now(), now(), NULL, '-2147483624', DEFAULT, DEFAULT, NULL, NULL) ON CONFLICT DO NOTHING;

-- notification_internal_user
INSERT INTO public.notification_internal_user
    (user_id, is_system, created_at, updated_at, deleted_at, resource_path)
VALUES('01H3K88JTY9P7QQBQ16Z63NZDK', true, now(), now(), NULL, '-2147483624') ON CONFLICT DO NOTHING;
-- resource_path -2147483629 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511RC1T3JCY', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483629', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511RC1T3JCY', '01GJZQ8C9TCT9H3VWAACGYE8YA', now(), now(), '-2147483629') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511RC1T3JCY', now(), now(), NULL, '-2147483629', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483630 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511REJ5S4VW', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483630', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511REJ5S4VW', '01GJZQ8C9TCT9H3VWAAFC4B05V', now(), now(), '-2147483630') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511REJ5S4VW', now(), now(), NULL, '-2147483630', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483631 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511RH54D63Y', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483631', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511RH54D63Y', '01GJZQ8C9TCT9H3VWAAFT4JYDK', now(), now(), '-2147483631') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511RH54D63Y', now(), now(), NULL, '-2147483631', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483632 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511RKX9BN53', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483632', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511RKX9BN53', '01GJZQ8C9TCT9H3VWAAK4VCS6K', now(), now(), '-2147483632') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511RKX9BN53', now(), now(), NULL, '-2147483632', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483633 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511RNH7T128', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483633', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511RNH7T128', '01GJZQ8C9TCT9H3VWAANCZJDWV', now(), now(), '-2147483633') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511RNH7T128', now(), now(), NULL, '-2147483633', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483634 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511RQ228NQA', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483634', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511RQ228NQA', '01GJZQ8C9TCT9H3VWAAR5ZXEQ2', now(), now(), '-2147483634') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511RQ228NQA', now(), now(), NULL, '-2147483634', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483635 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511RQAC7A1T', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483635', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511RQAC7A1T', '01GJZQ8C9TCT9H3VWAAS3XCJ2F', now(), now(), '-2147483635') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511RQAC7A1T', now(), now(), NULL, '-2147483635', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483637 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511RTAP0PFN', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483637', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511RTAP0PFN', '01GJZQ8C9TCT9H3VWAAWB9REXE', now(), now(), '-2147483637') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511RTAP0PFN', now(), now(), NULL, '-2147483637', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483638 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511RXKGKYNM', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483638', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511RXKGKYNM', '01GJZQ8C9TCT9H3VWAB03X8JP1', now(), now(), '-2147483638') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511RXKGKYNM', now(), now(), NULL, '-2147483638', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483639 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511RYY1KVSS', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483639', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511RYY1KVSS', '01GJZQ8C9TCT9H3VWAB346EBMR', now(), now(), '-2147483639') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511RYY1KVSS', now(), now(), NULL, '-2147483639', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483640 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511RYYT2CTR', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483640', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511RYYT2CTR', '01GJZQ8C9TCT9H3VWAB42HE3FA', now(), now(), '-2147483640') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511RYYT2CTR', now(), now(), NULL, '-2147483640', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483641 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511RZSK8KXQ', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483641', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511RZSK8KXQ', '01GJZQ8C9TCT9H3VWAB4PNWD3V', now(), now(), '-2147483641') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511RZSK8KXQ', now(), now(), NULL, '-2147483641', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483642 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511S3NMYZH1', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483642', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511S3NMYZH1', '01GJZQ8C9TCT9H3VWAB4YQG240', now(), now(), '-2147483642') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511S3NMYZH1', now(), now(), NULL, '-2147483642', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483643 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511S6Y4R1E5', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483643', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511S6Y4R1E5', '01GJZQ8C9TCT9H3VWAB55C8QG3', now(), now(), '-2147483643') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511S6Y4R1E5', now(), now(), NULL, '-2147483643', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483644 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511SA4RZ0EZ', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483644', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511SA4RZ0EZ', '01GJZQ8C9TCT9H3VWAB5X6AJJT', now(), now(), '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511SA4RZ0EZ', now(), now(), NULL, '-2147483644', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483645 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511SB4Q5GAN', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483645', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511SB4Q5GAN', '01GJZQ8C9TCT9H3VWAB9CW93S7', now(), now(), '-2147483645') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511SB4Q5GAN', now(), now(), NULL, '-2147483645', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483646 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511SDVG2G55', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483646', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511SDVG2G55', '01GJZQ8C9TCT9H3VWABAKDZG6Q', now(), now(), '-2147483646') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511SDVG2G55', now(), now(), NULL, '-2147483646', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483647 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511SHHHPBAG', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483647', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511SHHHPBAG', '01GJZQ8C9TCT9H3VWABDBRCRTD', now(), now(), '-2147483647') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511SHHHPBAG', now(), now(), NULL, '-2147483647', DEFAULT, DEFAULT, NULL, NULL);

-- resource_path -2147483648 --
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES('01GJZQC612DHFM7511SKKFK9RX', 'COUNTRY_JP', 'Usermgmt Schedule Job', '', NULL, 'schedule_job+usermgmt@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483648', true)
    ON CONFLICT DO NOTHING;

INSERT INTO user_group_member(
    user_id,
    user_group_id,
    created_at,
    updated_at,
    resource_path
) VALUES
    ('01GJZQC612DHFM7511SKKFK9RX', '01GJZQ8C9TCT9H3VWABFGP75A1', now(), now(), '-2147483648') ON CONFLICT DO NOTHING;

INSERT INTO staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES ('01GJZQC612DHFM7511SKKFK9RX', now(), now(), NULL, '-2147483648', DEFAULT, DEFAULT, NULL, NULL);
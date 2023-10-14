INSERT INTO public.schools
(school_id, name, country, city_id, district_id, point, is_system_school, created_at, updated_at, is_merge)
VALUES(-2147483644, 'End-to-end School', 'COUNTRY_JP', 1, 1, NULL, false, now(), now(), false)
ON CONFLICT DO NOTHING;

INSERT INTO public.users
(user_id, country, "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name)
VALUES('thu.vo+e2eadmin@manabie.com', 'COUNTRY_JP', 'thu.vo+e2eadmin@manabie.com', NULL, 'thu.vo+e2eadmin@manabie.com', 'thu.vo+e2eadmin@manabie.com', NULL, true, 'USER_GROUP_ADMIN', now(), now(), true, NULL, NULL, NULL, NULL, NULL, NULL)
ON CONFLICT DO NOTHING;

INSERT INTO public.users_groups
(user_id, group_id, is_origin, status, updated_at, created_at)
VALUES('thu.vo+e2eadmin@manabie.com', 'USER_GROUP_ADMIN', true, 'USER_GROUP_STATUS_ACTIVE', now(), now())
ON CONFLICT DO NOTHING;


INSERT INTO public.users
(user_id, country, "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name)
VALUES('thu.vo+e2eschool@manabie.com', 'COUNTRY_JP', 'thu.vo+e2eschool@manabie.com', '', 'thu.vo+e2eschool@manabie.com', 'thu.vo+e2eschool@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL)
ON CONFLICT DO NOTHING;

INSERT INTO public.users_groups
(user_id, group_id, is_origin, status, updated_at, created_at)
VALUES('thu.vo+e2eschool@manabie.com', 'USER_GROUP_SCHOOL_ADMIN', true, 'USER_GROUP_STATUS_ACTIVE', now(), now())
ON CONFLICT DO NOTHING;

INSERT INTO public.school_admins
(school_admin_id, school_id, updated_at, created_at)
VALUES('thu.vo+e2eschool@manabie.com',-2147483644, now(), now())
ON CONFLICT DO NOTHING;

\connect bob;

INSERT INTO public.school_level
(school_level_id, school_level_name, "sequence", is_archived, created_at, updated_at, deleted_at, resource_path)
VALUES('01G9KKDJA8QCDM9BWFNM5QTWSB', '[Aug 18] School Level 2', 2, false, '2022-08-04 11:56:59.976', '2022-08-18 11:27:04.619', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.school_level
(school_level_id, school_level_name, "sequence", is_archived, created_at, updated_at, deleted_at, resource_path)
VALUES('01G9KKDJA9AP1HAB1SRN6RYAZK', '[Aug 18] School Level 3', 3, false, '2022-08-04 11:56:59.977', '2022-08-18 11:27:04.619', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.school_level
(school_level_id, school_level_name, "sequence", is_archived, created_at, updated_at, deleted_at, resource_path)
VALUES('01GA5EPWQRHW0FK2VWKP4Z7FY3', '[Aug 18] School Level 4', 4, false, '2022-08-11 10:21:02.839', '2022-08-18 11:27:04.619', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.school_level
(school_level_id, school_level_name, "sequence", is_archived, created_at, updated_at, deleted_at, resource_path)
VALUES('01GAQK8TZTZKC6TPBDD1DAYJAM', '[Aug 18] School Level 5', 5, false, '2022-08-18 11:27:04.619', '2022-08-18 11:27:04.619', NULL, '-2147483644') ON CONFLICT DO NOTHING;


INSERT INTO public.school_info
(school_id, school_name, school_name_phonetic, is_archived, created_at, updated_at, deleted_at, resource_path, school_level_id, address, school_partner_id)
VALUES('01G9M1QKTC8N1XQY7GA11QYMAG', 'School Name 2', 'School Name Phonetic 2', false, '2022-08-04 16:07:09.260', '2022-08-04 16:07:09.260', NULL, '-2147483644', '01G9KKDJA9AP1HAB1SRN6RYAZK', '124 Nguyen Van Linh, Phuong 4, Quan 2, Thanh pho Ho Chi Minh', 'school_2') ON CONFLICT DO NOTHING;

INSERT INTO public.school_info
(school_id, school_name, school_name_phonetic, is_archived, created_at, updated_at, deleted_at, resource_path, school_level_id, address, school_partner_id)
VALUES('01G9M1QKTC8N1XQY7GA11QYMAJ', 'School Name 4', 'School Name Phonetic 4', false, '2022-08-04 16:07:09.260', '2022-08-04 16:07:09.260', NULL, '-2147483644', '01G9KKDJA8QCDM9BWFNM5QTWSB', '227 Nguyen Van Cu, Phuong 4, Quan 5, Thanh pho Ho Chi Minh', 'school_4') ON CONFLICT DO NOTHING;

INSERT INTO public.school_info
(school_id, school_name, school_name_phonetic, is_archived, created_at, updated_at, deleted_at, resource_path, school_level_id, address, school_partner_id)
VALUES('01GA5JPW36VSGZYXKB1NCVH5YZ', 'School 1', 'School 1', false, '2022-08-11 11:30:56.209', '2022-08-11 11:40:07.709', NULL, '-2147483644', '01GA5EPWQRHW0FK2VWKP4Z7FY3', '', 'school-1') ON CONFLICT DO NOTHING;   

INSERT INTO public.school_course
(school_course_id, school_course_name, school_course_name_phonetic, school_id, is_archived, created_at, updated_at, deleted_at, resource_path, school_course_partner_id)
VALUES('01G9RPYBFX6N81WB2K273916HT', 'English advanced speaking', 'English advanced speaking', '01G9M1QKTC8N1XQY7GA11QYMAG', false, '2022-08-06 11:34:47.933', '2022-08-06 11:34:47.933', NULL, '-2147483644', 'school_course_2') ON CONFLICT DO NOTHING;

INSERT INTO public.school_course
(school_course_id, school_course_name, school_course_name_phonetic, school_id, is_archived, created_at, updated_at, deleted_at, resource_path, school_course_partner_id)
VALUES('01GA5N7DP9639GSTG2NVR341WT', 'School Course 3', 'School Course 3', '01GA5JPW36VSGZYXKB1NCVH5YZ', false, '2022-08-11 12:14:55.600', '2022-08-11 12:16:57.850', NULL, '-2147483644', 'school_course_3') ON CONFLICT DO NOTHING;

INSERT INTO public.school_course
(school_course_id, school_course_name, school_course_name_phonetic, school_id, is_archived, created_at, updated_at, deleted_at, resource_path, school_course_partner_id)
VALUES('01GB7W3WT9AD8SQ9BTSJGQT16E', 'E2E - School Course 1661342969247-pakA9', 'E2E - School Course 1661342969247-pakA9', '01G9M1QKTC8N1XQY7GA11QYMAG', false, '2022-08-24 19:09:30.728', '2022-08-24 19:09:30.728', NULL, '-2147483644', 'e2e-school-course-1661342969247-pakA9') ON CONFLICT DO NOTHING;

INSERT INTO public.school_course
(school_course_id, school_course_name, school_course_name_phonetic, school_id, is_archived, created_at, updated_at, deleted_at, resource_path, school_course_partner_id)
VALUES('01GB7W3WT9AD8SQ9BTSJGQT16F', 'E2E - School Course 1661342969247-5ZGYC', 'E2E - School Course 1661342969247-5ZGYC', '01GA5JPW36VSGZYXKB1NCVH5YZ', false, '2022-08-24 19:09:30.728', '2022-08-24 19:09:30.728', NULL, '-2147483644', 'e2e-school-course-1661342969247-5ZGYC') ON CONFLICT DO NOTHING;

INSERT INTO public.school_course
(school_course_id, school_course_name, school_course_name_phonetic, school_id, is_archived, created_at, updated_at, deleted_at, resource_path, school_course_partner_id)
VALUES('01GB7W3WT9AD8SQ9BTSJGQT16H', 'E2E - School Course 1661342969247-eerJ0', 'E2E - School Course 1661342969247-eerJ0', '01G9M1QKTC8N1XQY7GA11QYMAJ', false, '2022-08-24 19:09:30.728', '2022-08-24 19:09:30.728', NULL, '-2147483644', 'e2e-school-course-1661342969247-eerJ0') ON CONFLICT DO NOTHING;


-- create student acc default: thu.vo+e2estudent@manabie.com/0UD6V1

INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name)
VALUES('01GV3FJ5X7J03N8GNJZVZC9PA3', 'COUNTRY_JP', 'vctqs1 student', NULL, NULL, 'thu.vo+e2estudent@manabie.com', NULL, false, 'USER_GROUP_STUDENT', '2023-03-09 22:24:18.833', '2023-03-09 22:24:18.833', false, NULL, NULL, false, false, NULL, NULL, '-2147483644', NULL, NULL, NULL, 'student', 'vctqs1', '', '', '', NULL, false, NULL, NULL) ON CONFLICT DO NOTHING;

INSERT INTO public.user_group
(user_group_id, user_group_name, created_at, updated_at, deleted_at, resource_path, org_location_id, is_system)
VALUES('01G4EQH8SDX8QZ8FXC771Q6H09', 'Student', '2023-03-09 17:52:48.097', '2023-03-09 17:52:48.097', NULL, '-2147483644', NULL, true) ON CONFLICT DO NOTHING;


INSERT INTO public.user_group_member
(user_id, user_group_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GV3FJ5X7J03N8GNJZVZC9PA3', '01G4EQH8SDX8QZ8FXC771Q6H09', '2023-03-09 22:24:18.878', '2023-03-09 22:24:18.878', NULL, '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.users_groups
(user_id, group_id, is_origin, status, updated_at, created_at, resource_path)
VALUES('01GV3FJ5X7J03N8GNJZVZC9PA3', 'USER_GROUP_STUDENT', true, 'USER_GROUP_STATUS_ACTIVE', '2023-03-09 22:24:18.876', '2023-03-09 22:24:18.876', '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.user_group
(user_group_id, user_group_name, created_at, updated_at, deleted_at, resource_path, org_location_id, is_system)
VALUES('01G4EQH8SDX8QZ8FXC771Q6H09', 'Student', '2023-03-09 17:52:48.097', '2023-03-09 17:52:48.097', NULL, '-2147483644', NULL, true) ON CONFLICT DO NOTHING;

INSERT INTO public.user_basic_info
(user_id, "name", first_name, last_name, full_name_phonetic, first_name_phonetic, last_name_phonetic, current_grade, grade_id, created_at, updated_at, deleted_at, resource_path, email)
VALUES('01GV3FJ5X7J03N8GNJZVZC9PA3', 'vctqs1 student', 'student', 'vctqs1', '', '', '', NULL, '01GB7EAYGAS312J0J2W87JR4N2', '2023-03-09 22:24:18.833', '2023-03-09 22:24:18.833', NULL, '-2147483644', 'thu.vo+e2estudent@manabie.com') ON CONFLICT DO NOTHING;
INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('01GV3FJ5X7J03N8GNJZVZC9PA3', '01GV2MBJRW9AS88X5S4C5DXCE7', NULL, '2023-03-09 22:24:18.813', '2023-03-09 22:24:18.813', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.user_address
(user_address_id, user_id, address_type, postal_code, prefecture_id, city, created_at, updated_at, deleted_at, resource_path, first_street, second_street)
VALUES('01GV3FJ63DDNKZ3YKEZK2J9XPA', '01GV3FJ5X7J03N8GNJZVZC9PA3', 'HOME_ADDRESS', '', NULL, '', '2023-03-09 22:24:18.925', '2023-03-09 22:24:18.925', NULL, '-2147483644', '', '') ON CONFLICT DO NOTHING;

INSERT INTO public.students
(student_id, current_grade, target_university, on_trial, billing_date, birthday, biography, updated_at, created_at, total_question_limit, school_id, deleted_at, additional_data, enrollment_status, resource_path, student_external_id, student_note, previous_grade, contact_preference, grade_id)
VALUES('01GV3FJ5X7J03N8GNJZVZC9PA3', NULL, NULL, true, '2023-03-09 22:24:18.879', NULL, NULL, '2023-03-09 22:24:18.879', '2023-03-09 22:24:18.879', 20, -2147483644, NULL, NULL, 'STUDENT_ENROLLMENT_STATUS_GRADUATED', '-2147483644', NULL, '', NULL, 'STUDENT_PHONE_NUMBER', '01GB7EAYGAS312J0J2W87JR4N2') ON CONFLICT DO NOTHING;


-- insert user_group is_system = false
INSERT INTO public.user_group
(user_group_id, user_group_name, created_at, updated_at, deleted_at, resource_path, org_location_id, is_system)
VALUES('01G5208C34W92QPK9MNY6ZWERJ', 'e2e-userGroup.1654703533783.0kZERp4CEe', '2022-06-08 22:52:14.180', '2022-06-08 22:52:14.180', NULL, '-2147483644', '01GV2MBJRW9AS88X5S4C5DXCE7', false) ON CONFLICT DO NOTHING;

INSERT INTO public.user_group
(user_group_id, user_group_name, created_at, updated_at, deleted_at, resource_path, org_location_id, is_system)
VALUES('01G5208PVBAVETP65QW95QAD6J', 'e2e-userGroup.1654703533825.eeb4hjJB9d', '2022-06-08 22:52:25.195', '2022-06-08 22:52:25.195', NULL, '-2147483644', '01GV2MBJRW9AS88X5S4C5DXCE7', false) ON CONFLICT DO NOTHING;



-- thu.vo+e2eteacher@manabie.com/123456789
INSERT INTO public.users
(user_id, country, "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name)
VALUES('01GV3GJC1M5QJGF2YB3TSYJ282', 'COUNTRY_JP', 'vctqs1 teacher', NULL, NULL, 'thu.vo+e2eteacher@manabie.com', NULL, NULL, 'USER_GROUP_TEACHER', '2023-03-09 22:41:53.662', '2023-03-09 22:41:53.662', NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483644', NULL, NULL, 'FEMALE', 'teacher', 'vctqs1', '', '', '', NULL, false, NULL, NULL) ON CONFLICT DO NOTHING;
INSERT INTO public.user_basic_info
(user_id, "name", first_name, last_name, full_name_phonetic, first_name_phonetic, last_name_phonetic, current_grade, grade_id, created_at, updated_at, deleted_at, resource_path, email)
VALUES('01GV3GJC1M5QJGF2YB3TSYJ282', 'vctqs1 teacher', 'teacher', 'vctqs1', '', '', '', NULL, NULL, '2023-03-09 22:41:53.662', '2023-03-09 22:41:53.662', NULL, '-2147483644', 'thu.vo+e2eteacher@manabie.com') ON CONFLICT DO NOTHING;
INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('01GV3GJC1M5QJGF2YB3TSYJ282', '01GV2MBJV336BV73298ED0ND2X', NULL, '2023-03-09 22:41:53.669', '2023-03-09 22:41:53.669', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('01GV3GJC1M5QJGF2YB3TSYJ282', '01GV2MBJX12HR7N1CTF76CR7VH', NULL, '2023-03-09 22:41:53.669', '2023-03-09 22:41:53.669', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('01GV3GJC1M5QJGF2YB3TSYJ282', '01GV2MBJRW9AS88X5S4C5DXCE7', NULL, '2023-03-09 22:41:53.669', '2023-03-09 22:41:53.669', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.teachers
(teacher_id, school_ids, updated_at, created_at, school_name, deleted_at, resource_path)
VALUES('01GV3GJC1M5QJGF2YB3TSYJ282', '{-2147483644}', '2023-03-09 22:41:56.492', '2023-03-09 22:41:56.492', NULL, NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.usr_email
(email, usr_id, create_at, updated_at, delete_at, resource_path, import_id)
VALUES('thu.vo+e2eteacher@manabie.com', '01GV3GJC1M5QJGF2YB3TSYJ282', '2023-03-09 22:41:53.588', '2023-03-09 22:41:53.588', NULL, '-2147483644', 3519810662228296706) ON CONFLICT DO NOTHING;
INSERT INTO public.users_groups
(user_id, group_id, is_origin, status, updated_at, created_at, resource_path)
VALUES('01GV3GJC1M5QJGF2YB3TSYJ282', 'USER_GROUP_TEACHER', true, 'USER_GROUP_STATUS_ACTIVE', '2023-03-09 22:41:53.662', '2023-03-09 22:41:53.662', '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.user_group_member
(user_id, user_group_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GV3GJC1M5QJGF2YB3TSYJ282', '01G5208PVBAVETP65QW95QAD6J', '2023-03-09 23:03:17.895', '2023-03-09 23:03:17.895', NULL, '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES('01GV3GJC1M5QJGF2YB3TSYJ282', '2023-03-08 01:24:01.765', '2023-03-08 01:24:01.765', NULL, '-2147483644', false, 'AVAILABLE', '2023-03-07', '2023-03-07');


INSERT INTO public.granted_role
(granted_role_id, user_group_id, role_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GHG5K0WF02AC5DTV651RMYJ2', '01G5208C34W92QPK9MNY6ZWERJ', '01G1GQEKEHXKSM78NBW96NJ7H8', '2022-11-10 14:03:09.967', '2022-11-10 14:03:09.967', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role
(granted_role_id, user_group_id, role_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GHG5MSCY4FW4BM7EG7TJGD2A', '01G5208PVBAVETP65QW95QAD6J', '01G1GQEKEHXKSM78NBW96NJ7H8', '2022-11-10 14:04:07.838', '2022-11-10 14:04:13.411', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role
(granted_role_id, user_group_id, role_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GHG5MYV1NAE59MXGK04DD5A8', '01G4EQH8SDX8QZ8FXC771Q6H09', '01G1GQEKEHXKSM78NBW96NJ7H9', '2022-11-10 14:04:13.415', '2022-11-10 14:04:13.415', NULL, '-2147483644') ON CONFLICT DO NOTHING;




INSERT INTO public.granted_role_access_path
(granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GVMMV6T0P6ETC1QJXWBT4JEH', '01GV2MBK3ZCQEJ711H3EW0VEQY', '2022-08-02 18:07:02.938', '2022-08-02 18:07:02.938', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role_access_path
(granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GHG5K0WF02AC5DTV651RMYJ2', '01GV2MBJRW9AS88X5S4C5DXCE7', '2022-08-02 18:07:02.938', '2022-08-02 18:07:02.938', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role_access_path
(granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GHG5MSCY4FW4BM7EG7TJGD2A', '01GV2MBJR5PGRHC601V0NMHKGY', '2022-07-26 15:54:40.351', '2022-07-26 15:54:40.351', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role_access_path
(granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GHG5MYV1NAE59MXGK04DD5A8', '01GV2MBJV336BV73298ED0ND2X', '2022-07-27 10:08:44.003', '2022-07-27 10:08:44.003', NULL, '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.granted_role_access_path
(granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GHG5MYV1NAE59MXGK04DD5A8', '01GV2MBJX12HR7N1CTF76CR7VH', '2022-08-02 18:07:02.938', '2022-08-02 18:07:02.938', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role_access_path
(granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GHG5MSCY4FW4BM7EG7TJGD2A', '01GV2MBK19NHHG2PV1SB4JHTEH', '2022-07-26 16:04:38.332', '2022-07-26 16:04:38.332', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role_access_path
(granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GHG5K0WF02AC5DTV651RMYJ2', '01GV2MBK3ZCQEJ711H3EW0VEQY', '2022-07-27 10:08:44.003', '2022-07-27 10:08:44.003', NULL, '-2147483644') ON CONFLICT DO NOTHING;


INSERT INTO public.granted_role_access_path
(granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GHG5MYV1NAE59MXGK04DD5A8', '01GV2MBK4PP8M0V7JB4WGR315G', '2022-08-02 18:07:02.938', '2022-08-02 18:07:02.938', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role_access_path
(granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GHG5K0WF02AC5DTV651RMYJ2', '01GV2MBK66A6SY7V0TVA8KWJHG', '2022-07-26 15:54:40.351', '2022-07-26 15:54:40.351', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role_access_path
(granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GHG5MSCY4FW4BM7EG7TJGD2A', '01GV2MBK805A0QREYJSKS92R6A', '2022-07-26 16:04:38.332', '2022-07-26 16:04:38.332', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role_access_path
(granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GHG5MYV1NAE59MXGK04DD5A8', '01GV2MBK94A3Q1184KK8KQJCB7', '2022-07-27 10:08:44.003', '2022-07-27 10:08:44.003', NULL, '-2147483644') ON CONFLICT DO NOTHING;
INSERT INTO public.granted_role_access_path
(granted_role_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GHG5MSCY4FW4BM7EG7TJGD2A', '01GV2MBKA50D3ZY095W69RCX0S', '2022-07-27 10:08:44.003', '2022-07-27 10:08:44.003', NULL, '-2147483644') ON CONFLICT DO NOTHING;



INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('thu.vo+e2eschool@manabie.com', '01GV2MBJR5PGRHC601V0NMHKGY', NULL, '2023-03-09 23:50:42.195', '2023-03-09 23:50:42.195', NULL, '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('thu.vo+e2eschool@manabie.com', '01GV2MBJRW9AS88X5S4C5DXCE7', NULL, '2023-03-09 23:50:42.195', '2023-03-09 23:50:42.195', NULL, '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('thu.vo+e2eschool@manabie.com', '01GV2MBJRW9AS88X5S4C5DXCE7', NULL, '2023-03-09 23:50:42.195', '2023-03-09 23:50:42.195', NULL, '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('thu.vo+e2eschool@manabie.com', '01GV2MBJV336BV73298ED0ND2X', NULL, '2023-03-09 23:50:42.195', '2023-03-09 23:50:42.195', NULL, '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('thu.vo+e2eschool@manabie.com', '01GV2MBJX12HR7N1CTF76CR7VH', NULL, '2023-03-09 23:50:42.195', '2023-03-09 23:50:42.195', NULL, '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('thu.vo+e2eschool@manabie.com', '01GV2MBJZW0CA66KXAHJTZJ86X', NULL, '2023-03-09 23:50:42.195', '2023-03-09 23:50:42.195', NULL, '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('thu.vo+e2eschool@manabie.com', '01GV2MBK19NHHG2PV1SB4JHTEH', NULL, '2023-03-09 23:50:42.195', '2023-03-09 23:50:42.195', NULL, '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('thu.vo+e2eschool@manabie.com', '01GV2MBK3ZCQEJ711H3EW0VEQY', NULL, '2023-03-09 23:50:42.195', '2023-03-09 23:50:42.195', NULL, '-2147483644') ON CONFLICT DO NOTHING;

--

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('thu.vo+e2eschool@manabie.com', '01GV2MBK4PP8M0V7JB4WGR315G', NULL, '2023-03-09 23:50:42.195', '2023-03-09 23:50:42.195', NULL, '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('thu.vo+e2eschool@manabie.com', '01GV2MBK66A6SY7V0TVA8KWJHG', NULL, '2023-03-09 23:50:42.195', '2023-03-09 23:50:42.195', NULL, '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('thu.vo+e2eschool@manabie.com', '01GV2MBK805A0QREYJSKS92R6A', NULL, '2023-03-09 23:50:42.195', '2023-03-09 23:50:42.195', NULL, '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('thu.vo+e2eschool@manabie.com', '01GV2MBK94A3Q1184KK8KQJCB7', NULL, '2023-03-09 23:50:42.195', '2023-03-09 23:50:42.195', NULL, '-2147483644') ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('thu.vo+e2eschool@manabie.com', '01GV2MBKA50D3ZY095W69RCX0S', NULL, '2023-03-09 23:50:42.195', '2023-03-09 23:50:42.195', NULL, '-2147483644') ON CONFLICT DO NOTHING;


-- create phuc.chau+e2ehcmschooladmin@manabie.com	 in e2e-hcm tenant
 
INSERT INTO public.organizations
(organization_id, tenant_id, name, resource_path, domain_name, logo_url, country, created_at, updated_at, deleted_at, scrypt_signer_key, scrypt_salt_separator, scrypt_rounds, scrypt_memory_cost)
VALUES('-2147483638', 'e2e-hcm-joddm', 'E2E HCM', '-2147483638', 'e2e-hcm', 'https://storage.googleapis.com/prod-tokyo-backend/user-upload/manabie_ic_splash.png', 'COUNTRY_JP', '2022-06-24 18:11:22.153', '2022-06-24 18:11:22.153', NULL, NULL, NULL, NULL, NULL)
 ON CONFLICT(organization_id) DO UPDATE SET tenant_id = 'e2e-hcm-joddm', domain_name = 'e2e-hcm';

update organization_auths set auth_project_id='dev-manabie-online', auth_tenant_id='e2e-hcm-joddm' where organization_id='-2147483638';

INSERT INTO public.users
(user_id, resource_path, country, "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name)
VALUES('6a4NZjhmpEZ4b93it9HE1EbjtBC2', '-2147483638', 'COUNTRY_JP', 'phuc.chau+e2ehcmschooladmin@manabie.com', '', 'phuc.chau+e2ehcmschooladmin@manabie.com', 'phuc.chau+e2ehcmschooladmin@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL)
ON CONFLICT DO NOTHING;

INSERT INTO public.users_groups
(user_id, resource_path, group_id, is_origin, status, updated_at, created_at)
VALUES('6a4NZjhmpEZ4b93it9HE1EbjtBC2', '-2147483638', 'USER_GROUP_SCHOOL_ADMIN', true, 'USER_GROUP_STATUS_ACTIVE', now(), now())
ON CONFLICT DO NOTHING;

INSERT INTO public.school_admins
(school_admin_id, resource_path, school_id, updated_at, created_at)
VALUES('6a4NZjhmpEZ4b93it9HE1EbjtBC2', '-2147483638',-2147483638, now(), now())
ON CONFLICT DO NOTHING;

INSERT INTO public.user_group
  (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
  ('01GE1NVHERDCK5CBAW1SP80SPQ_2147483638', 'Teacher', true, now(), now(), '-2147483638'),
  ('01GE1NVHERDCK5CBAW1P99AVZ3_2147483638', 'School Admin', true, now(), now(), '-2147483638'),
  ('01GE1NVHERDCK5CBAW1MNEVRZE_2147483638', 'HQ Staff', true, now(), now(), '-2147483638'),
  ('01GE1NVHERDCK5CBAW1GW3T59Y_2147483638', 'Centre Lead', true, now(), now(), '-2147483638'),
  ('01GE1NVHERDCK5CBAW1EE6Y4AC_2147483638', 'Teacher Lead', true, now(), now(), '-2147483638'),
  ('01GE1NVHERDCK5CBAW1DP32TGW_2147483638', 'Centre Manager', true, now(), now(), '-2147483638'),
  ('01GE1NVHERDCK5CBAW1BF9DDDX_2147483638', 'Centre Staff', true, now(), now(), '-2147483638')
  ON CONFLICT DO NOTHING;

INSERT INTO public.granted_role
  (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01GE1NVHERDCK5CBAW1SP80SPQ_2147483638', '01GE1NVHERDCK5CBAW1SP80SPQ_2147483638', '01G6A6B3YP6P161SH6CJN3QPF0', now(), now(), '-2147483638'),
  ('01GE1NVHERDCK5CBAW1P99AVZ3_2147483638', '01GE1NVHERDCK5CBAW1P99AVZ3_2147483638', '01G6A6B3YP6P161SH6CJN3QPF1', now(), now(), '-2147483638')

  ON CONFLICT DO NOTHING;


  INSERT INTO public.location_types
  (location_type_id, "name", display_name, parent_name, parent_location_type_id, updated_at, created_at, deleted_at, resource_path, is_archived, "level")
  VALUES('01GV2MB7J1XH9JN4RGA2Y06JZC_2147483638', 'brand', 'Brand - E2E HCM', 'org', '01FR4M51XJY9E77GSN4QZ1Q8M2', '2023-03-09 14:28:39.679', '2023-03-09 14:28:39.489', NULL, '-2147483638', false, 1) ON CONFLICT DO NOTHING;
  INSERT INTO public.location_types
  (location_type_id, "name", display_name, parent_name, parent_location_type_id, updated_at, created_at, deleted_at, resource_path, is_archived, "level")
  VALUES('01GV2MB7K7P24Q80HWSXNB0RN8_2147483638', 'center', 'Center - E2E HCM', 'brand', '01GV2MB7J1XH9JN4RGA2Y06JZC_2147483638', '2023-03-09 14:28:39.679', '2023-03-09 14:28:39.527', NULL, '-2147483638', false, 2) ON CONFLICT DO NOTHING;



  INSERT INTO public.locations
  (location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
  VALUES('01GV2MBJR5PGRHC601V0NMHKGY_2147483638', 'Brand 1 - E2E HCM', '2023-03-09 14:28:50.933', '2023-03-09 14:28:51.563', NULL, '-2147483638', '01GV2MB7J1XH9JN4RGA2Y06JZC_2147483638', '1', NULL, '01FR4M51XJY9E77GSN4QZ1Q8N2', false, '01FR4M51XJY9E77GSN4QZ1Q8N2/01GV2MBJR5PGRHC601V0NMHKGY_2147483638') ON CONFLICT DO NOTHING;
  INSERT INTO public.locations
  (location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
  VALUES('01GV2MBJRW9AS88X5S4C5DXCE7_2147483638', 'Center 2 - E2E HCM', '2023-03-09 14:28:50.933', '2023-03-09 14:28:52.563', NULL, '-2147483638', '01GV2MB7K7P24Q80HWSXNB0RN8_2147483638', '2', '1', '01GV2MBJR5PGRHC601V0NMHKGY_2147483638', false, '01FR4M51XJY9E77GSN4QZ1Q8N2/01GV2MBJR5PGRHC601V0NMHKGY_2147483638/01GV2MBJRW9AS88X5S4C5DXCE7_2147483638') ON CONFLICT DO NOTHING;


INSERT INTO public.granted_role_access_path
  (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
  ('01GE1NVHERDCK5CBAW1SP80SPQ_2147483638', '01GV2MBJR5PGRHC601V0NMHKGY_2147483638', now(), now(), '-2147483638'),
  ('01GE1NVHERDCK5CBAW1P99AVZ3_2147483638', '01GV2MBJR5PGRHC601V0NMHKGY_2147483638', now(), now(), '-2147483638'),
  ('01GE1NVHERDCK5CBAW1SP80SPQ_2147483638', '01GV2MBJRW9AS88X5S4C5DXCE7_2147483638', now(), now(), '-2147483638'),
  ('01GE1NVHERDCK5CBAW1P99AVZ3_2147483638', '01GV2MBJRW9AS88X5S4C5DXCE7_2147483638', now(), now(), '-2147483638'),
  ('01GE1NVHERDCK5CBAW1SP80SPQ_2147483638', '01FR4M51XJY9E77GSN4QZ1Q8N2', now(), now(), '-2147483638'),
  ('01GE1NVHERDCK5CBAW1P99AVZ3_2147483638', '01FR4M51XJY9E77GSN4QZ1Q8N2', now(), now(), '-2147483638')
  ON CONFLICT DO NOTHING;

INSERT INTO user_group_member 
  (user_id, user_group_id, created_at, updated_at, deleted_at, resource_path) 
VALUES 
  ('6a4NZjhmpEZ4b93it9HE1EbjtBC2', '01GE1NVHERDCK5CBAW1P99AVZ3_2147483638', now(), now(), null, '-2147483638') 
  ON CONFLICT DO NOTHING;

-- create school admin account for managara-hs
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name)
VALUES('FDirGMMdFtUxxO5sd6KEW6mjc5n1', 'COUNTRY_JP', 'School Admin Managara HS', NULL, NULL, 'schooladmin_managara-hs@manabie.com', NULL, false, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), false, NULL, NULL, false, false, NULL, NULL, '-2147483629', NULL, NULL, NULL, 'Managara HS', 'School Admin', '', '', '', NULL, false, NULL, NULL) ON CONFLICT DO NOTHING;

INSERT INTO public.staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES('FDirGMMdFtUxxO5sd6KEW6mjc5n1', now(), now(), NULL, '-2147483629', false, 'AVAILABLE', '2023-03-01', '2099-03-01') ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('FDirGMMdFtUxxO5sd6KEW6mjc5n1', '01GFMNHQ1WHGRC8AW6K913AM3G', NULL, now(), now(), NULL, '-2147483629') ON CONFLICT DO NOTHING;

INSERT INTO public.user_group_member
(user_id, user_group_id, created_at, updated_at, deleted_at, resource_path)
VALUES('FDirGMMdFtUxxO5sd6KEW6mjc5n1', '01GFMNHQ1ZY20VNJCJZQ5EX4M2', now(), now(), NULL, '-2147483629') ON CONFLICT DO NOTHING;

-- create school admin account for managara-base
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name)
VALUES('umFwXlyeshaCWjfLDKR71o1pjyf1', 'COUNTRY_JP', 'School Admin Managara Base', NULL, NULL, 'schooladmin_managara-base@manabie.com', NULL, false, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), false, NULL, NULL, false, false, NULL, NULL, '-2147483630', NULL, NULL, NULL, 'Managara Base', 'School Admin', '', '', '', NULL, false, NULL, NULL) ON CONFLICT DO NOTHING;

INSERT INTO public.staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES('umFwXlyeshaCWjfLDKR71o1pjyf1', now(), now(), NULL, '-2147483630', false, 'AVAILABLE', '2023-03-01', '2099-03-01') ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('umFwXlyeshaCWjfLDKR71o1pjyf1', '01GFMMFRXC6SKTTT44HWR3BRY8', NULL, now(), now(), NULL, '-2147483630') ON CONFLICT DO NOTHING;

INSERT INTO public.user_group_member
(user_id, user_group_id, created_at, updated_at, deleted_at, resource_path)
VALUES('umFwXlyeshaCWjfLDKR71o1pjyf1', '01GFMMFRXD7KWYFJ6MG22M62M2', now(), now(), NULL, '-2147483630') ON CONFLICT DO NOTHING;

-- create school admin account for manabie
INSERT INTO public.users
(user_id, country, name, avatar, phone_number, username, login_email, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, last_login_date, birthday, gender, first_name, last_name, first_name_phonetic, last_name_phonetic, full_name_phonetic, remarks, is_system, user_external_id, previous_name)
VALUES('9dc5jYe5m1hfhiGeBmEXH74rwKF3', 'COUNTRY_JP', 'School Admin Manabie', NULL, NULL, 'schooladmin_managara-manabie@manabie.com','schooladmin_managara-manabie@manabie.com',  'schooladmin_managara-manabie@manabie.com', NULL, false, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), false, NULL, NULL, false, false, NULL, NULL, '-2147483648', NULL, NULL, NULL, 'Manabie', 'School Admin', '', '', '', NULL, false, NULL, NULL) ON CONFLICT DO NOTHING;

INSERT INTO public.staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES('9dc5jYe5m1hfhiGeBmEXH74rwKF3', now(), now(), NULL, '-2147483648', false, 'AVAILABLE', '2023-03-01', '2099-03-01') ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES('9dc5jYe5m1hfhiGeBmEXH74rwKF3', '01FR4M51XJY9E77GSN4QZ1Q9N1', NULL, now(), now(), NULL, '-2147483648') ON CONFLICT DO NOTHING;

INSERT INTO public.user_group_member
(user_id, user_group_id, created_at, updated_at, deleted_at, resource_path)
VALUES('9dc5jYe5m1hfhiGeBmEXH74rwKF3', '01G4EQH8SDX8QZ8FXC771Q1H01', now(), now(), NULL, '-2147483648') ON CONFLICT DO NOTHING;


-- insert data for org e2e
INSERT INTO public.staff
(staff_id, created_at, updated_at, deleted_at, resource_path, auto_create_timesheet, working_status, start_date, end_date)
VALUES('thu.vo+e2eschool@manabie.com', now(), now(), NULL, '-2147483644', false, 'AVAILABLE', '2023-03-01', '2099-03-01') ON CONFLICT DO NOTHING;


INSERT INTO public.users
(user_id, country, "name", avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES
	('bdd_admin-manabie', 'COUNTRY_JP', 'bdd_admin-manabie@manabie.com', '', 'bdd_admin-manabie@manabie.com', 'bdd_admin-manabie@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483648', true),
	('bdd_admin-jprep', 'COUNTRY_JP', 'bdd_admin-jprep@manabie.com', '', 'bdd_admin-jprep@manabie.com', 'bdd_admin-jprep@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483647', true),
	('bdd_admin-e2e', 'COUNTRY_JP', 'bdd_admin-e2e@manabie.com', '', 'bdd_admin-e2e@manabie.com', 'bdd_admin-e2e@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483644', true),
	('bdd_admin-kec-demo', 'COUNTRY_JP', 'bdd_admin-kec-demo@manabie.com', '', 'bdd_admin-kec-demo@manabie.com', 'bdd_admin-kec-demo@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483635', true),
	('bdd_admin-managara-base', 'COUNTRY_JP', 'bdd_admin-managara-base@manabie.com', '', 'bdd_admin-managara-base@manabie.com', 'bdd_admin-managara-base@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483630', true),
	('bdd_admin-managara-hs', 'COUNTRY_JP', 'bdd_admin-managara-hs@manabie.com', '', 'bdd_admin-managara-hs@manabie.com', 'bdd_admin-managara-hs@manabie.com', NULL, NULL, 'USER_GROUP_SCHOOL_ADMIN', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483629', true)
	ON CONFLICT DO NOTHING;

-- user_group_member
INSERT INTO public.user_group_member(
	user_id, user_group_id, created_at, updated_at, deleted_at, resource_path)
VALUES 
	('bdd_admin-manabie', '01G4EQH8SDX8QZ8FXC771Q1H01', now(), now(), null, '-2147483648'),
	('bdd_admin-jprep', '01G4EQH8SDX8QZ8FXC771Q1H02', now(), now(), null, '-2147483647'),
	('bdd_admin-e2e', '01G4EQH8SDX8QZ8FXC771Q1H05', now(), now(), null, '-2147483644'),
	('bdd_admin-kec-demo', '01G8T4A85RX6FPW6A74SVSWZ4J2', now(), now(), null, '-2147483635'),
	('bdd_admin-managara-base', '01GFMMFRXD7KWYFJ6MG22M62M2', now(), now(), null, '-2147483630'),
	('bdd_admin-managara-hs', '01GFMNHQ1ZY20VNJCJZQ5EX4M2', now(), now(), null, '-2147483629')
	ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths(
	user_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
VALUES
	('bdd_admin-manabie', '01FR4M51XJY9E77GSN4QZ1Q9N1', null, now(), now(), null, '-2147483648'),
	('bdd_admin-jprep', '01FR4M51XJY9E77GSN4QZ1Q9N2', null, now(), now(), null, '-2147483647'),
	('bdd_admin-e2e', '01FR4M51XJY9E77GSN4QZ1Q9N5', null, now(), now(), null, '-2147483644'),
	('bdd_admin-kec-demo', '01FR4M51XJY9E77GSN4QZ1Q8N4', null, now(), now(), null, '-2147483635'),
	('bdd_admin-managara-base', '01GFMMFRXC6SKTTT44HWR3BRY8', null, now(), now(), null, '-2147483630'),
	('bdd_admin-managara-hs', '01GFMNHQ1WHGRC8AW6K913AM3G', null, now(), now(), null, '-2147483629')
	ON CONFLICT DO NOTHING;

update organizations set tenant_id = 'manabie-0nl6t', domain_name= 'manabie' where resource_path = '-2147483648';
update organizations set tenant_id = 'end-to-end-dopvo', domain_name= 'e2e' where resource_path = '-2147483644';
update organizations set tenant_id = 'kec-demo-4ybj3', domain_name= 'kec-demo' where resource_path = '-2147483635';
update organizations set tenant_id = 'withus-managara-base-0wf23', domain_name = 'managara-base' where resource_path = '-2147483630';
update organizations set tenant_id = 'withus-managara-hs-t5fuk', domain_name= 'managara-hs' where resource_path = '-2147483629';

update organization_auths set auth_tenant_id = 'withus-managara-base-0wf23', auth_project_id = 'dev-manabie-online' where organization_id = '-2147483630';
update organization_auths set auth_tenant_id = 'withus-managara-hs-t5fuk', auth_project_id = 'dev-manabie-online' where organization_id = '-2147483629';
update organization_auths set auth_tenant_id = 'manabie-0nl6t', auth_project_id = 'dev-manabie-online' where organization_id = '-2147483648';
update organization_auths set auth_tenant_id = 'kec-demo-4ybj3', auth_project_id = 'dev-manabie-online' where organization_id = '-2147483635';
INSERT INTO public.organization_auths
  (organization_id, auth_project_id, auth_tenant_id)
VALUES
  (-2147483644, 'dev-manabie-online', 'end-to-end-dopvo'),
  (-2147483648, 'fake_aud', 'manabie-0nl6t'),
  (-2147483635, 'fake_aud', 'kec-demo-4ybj3')
  ON CONFLICT DO NOTHING;

-- insert master.location.write for managara-base
INSERT INTO permission
  (permission_id, permission_name, created_at, updated_at, resource_path)
VALUES 
  ('01GM2DS7WP4S4JY7Q68P78NVZ8', 'master.location.write', NOW(), NOW(), '-2147483630')
  ON CONFLICT DO NOTHING;

with 
role as (
  select r.role_id, r.resource_path
  from "role" r
  where r.role_name = ANY ('{School Admin}')
	and r.resource_path = ANY ('{-2147483630}')),
permission as (
	select p.permission_id, p.resource_path
  from "permission" p 
  where p.permission_name = ANY ('{
	master.location.write}')
	and p.resource_path = ANY ('{-2147483630}'))

insert into permission_role
  (permission_id, role_id, created_at, updated_at, resource_path)
select permission.permission_id,role.role_id, now(), now(), role.resource_path
  from role, permission
  where role.resource_path = permission.resource_path
  on conflict on constraint permission_role__pk do nothing;

INSERT INTO user_tag (user_tag_id,user_tag_name,user_tag_type,is_archived,created_at,updated_at,deleted_at,resource_path,user_tag_partner_id) VALUES
  ('01GWK65MPAN75Q1928GNRP07FT','Parent Tag Discount 1','USER_TAG_TYPE_PARENT_DISCOUNT',false,now(),now(),NULL,'-2147483630','parent_tag_discount_1'),
  ('01GWK65MPAN75Q1928GNRP07FW','Student Tag 3','USER_TAG_TYPE_STUDENT',false,now(),now(),NULL,'-2147483630','4_student_tag_3'),
  ('01GWK65MP9W0N3CG51SXT9A5V7','Student Tag 1','USER_TAG_TYPE_STUDENT',false,now(),now(),NULL,'-2147483630','2_student_tag_1'),
  ('01GWK65MPAN75Q1928GNRP07FV','Student Tag 2','USER_TAG_TYPE_STUDENT',false,now(),now(),NULL,'-2147483630','3_student_tag_2'),
  ('01GWK65MPAN75Q1928GNRP07FR','Parent Tag 1','USER_TAG_TYPE_PARENT',false,now(),now(),NULL,'-2147483630','parent_tag_1'),
  ('01GWK65MPAN75Q1928GNRP07FS','Student Tag Discount 1','USER_TAG_TYPE_STUDENT_DISCOUNT',false,now(),now(),NULL,'-2147483630','student_tag_discount_1'),
  ('01GWK3TP8QD0NFHP9C680EPD7J','Student Tag 2','USER_TAG_TYPE_STUDENT',false,now(),now(),NULL,'-2147483629','3_student_tag_2'),
  ('01GWK3TP8QD0NFHP9C680EPD7F','Parent Tag 1','USER_TAG_TYPE_PARENT',false,now(),now(),NULL,'-2147483629','parent_tag_1'),
  ('01GWK3TP8QD0NFHP9C680EPD7G','Student Tag Discount 1','USER_TAG_TYPE_STUDENT_DISCOUNT',false,now(),now(),NULL,'-2147483629','student_tag_discount_1'),
  ('01GWK3TP8QD0NFHP9C680EPD7H','Parent Tag Discount 1','USER_TAG_TYPE_PARENT_DISCOUNT',false,now(),now(),NULL,'-2147483629','parent_tag_discount_1'),
  ('01GWK3TP8QD0NFHP9C680EPD7K','Student Tag 3','USER_TAG_TYPE_STUDENT',false,now(),now(),NULL,'-2147483629','4_student_tag_3'),
  ('01GWK3TP8QD0NFHP9C680EPD7E','Student Tag 1','USER_TAG_TYPE_STUDENT',false,now(),now(),NULL,'-2147483629','2_student_tag_1'),
  ('-2147483648tag_user_student_01', '[Manabie]Tag User Student 1', 'USER_TAG_TYPE_STUDENT', false, now(),  now(), NULL, '-2147483648', 'tag_user_student_partner_01'),
  ('-2147483648tag_user_student_02', '[Manabie]Tag User Student 2', 'USER_TAG_TYPE_STUDENT', false,  now(),  now(), NULL, '-2147483648', 'tag_user_student_partner_02'),
  ('-2147483648tag_user_student_discount_01', '[Manabie]Tag User Discount Student 1', 'USER_TAG_TYPE_STUDENT_DISCOUNT', false,  now(),  now(), NULL, '-2147483648', 'tag_user_student_discount_partner_01'),
  ('-2147483648tag_user_student_discount_02', '[Manabie]Tag User Discount Student 2', 'USER_TAG_TYPE_STUDENT_DISCOUNT', false,  now(),  now(), NULL, '-2147483648', 'tag_user_student_discount_partner_02'),
  ('-2147483648tag_user_parent_01', '[Manabie]Tag User Parent 1', 'USER_TAG_TYPE_PARENT', false,  now(),  now(), NULL, '-2147483648', 'tag_user_parent_partner_01'),
  ('-2147483648tag_user_parent_02', '[Manabie]Tag User Parent 2', 'USER_TAG_TYPE_PARENT', false,  now(),  now(), NULL, '-2147483648', 'tag_user_parent_partner_02'),
  ('-2147483648tag_user_parent_discount_01', '[Manabie]Tag User Discount Parent 1', 'USER_TAG_TYPE_PARENT_DISCOUNT', false,  now(),  now(), NULL, '-2147483648', 'tag_user_parent_discount_partner_01'),
  ('-2147483648tag_user_parent_discount_02', '[Manabie]Tag User Discount Parent 2', 'USER_TAG_TYPE_PARENT_DISCOUNT', false,  now(),  now(), NULL, '-2147483648', 'tag_user_parent_discount_partner_02'),
  ('-2147483635tag_user_student_01', '[kec-demo]Tag User Student 1', 'USER_TAG_TYPE_STUDENT', false, now(),  now(), NULL, '-2147483635', 'tag_user_student_partner_01'),
  ('-2147483635tag_user_student_02', '[kec-demo]Tag User Student 2', 'USER_TAG_TYPE_STUDENT', false,  now(),  now(), NULL, '-2147483635', 'tag_user_student_partner_02'),
  ('-2147483635tag_user_student_discount_01', '[kec-demo]Tag User Discount Student 1', 'USER_TAG_TYPE_STUDENT_DISCOUNT', false,  now(),  now(), NULL, '-2147483635', 'tag_user_student_discount_partner_01'),
  ('-2147483635tag_user_student_discount_02', '[kec-demo]Tag User Discount Student 2', 'USER_TAG_TYPE_STUDENT_DISCOUNT', false,  now(),  now(), NULL, '-2147483635', 'tag_user_student_discount_partner_02'),
  ('-2147483635tag_user_parent_01', '[kec-demo]Tag User Parent 1', 'USER_TAG_TYPE_PARENT', false,  now(),  now(), NULL, '-2147483635', 'tag_user_parent_partner_01'),
  ('-2147483635tag_user_parent_02', '[kec-demo]Tag User Parent 2', 'USER_TAG_TYPE_PARENT', false,  now(),  now(), NULL, '-2147483635', 'tag_user_parent_partner_02'),
  ('-2147483635tag_user_parent_discount_01', '[kec-demo]Tag User Discount Parent 1', 'USER_TAG_TYPE_PARENT_DISCOUNT', false,  now(),  now(), NULL, '-2147483635', 'tag_user_parent_discount_partner_01'),
  ('-2147483635tag_user_parent_discount_02', '[kec-demo]Tag User Discount Parent 2', 'USER_TAG_TYPE_PARENT_DISCOUNT', false,  now(),  now(), NULL, '-2147483635', 'tag_user_parent_discount_partner_02')
ON CONFLICT DO NOTHING;

INSERT INTO location_types (location_type_id,"name",display_name,parent_name,parent_location_type_id,updated_at,created_at,deleted_at,resource_path,is_archived,"level") VALUES
  ('01GWK38KV1NZ7654N56B6WNVJ4','center','center','brand','01GWK38KV0M6TM5869EH9VFEHC',now(),now(),NULL,'-2147483629',false,2),
  ('01GWK38KV0M6TM5869EH9VFEHC','brand','brand','org','01GFMNHQ1WHGRC8AW6K913AM3G',now(),now(),NULL,'-2147483629',false,1)
  ON CONFLICT DO NOTHING;

INSERT INTO locations (location_id,"name",created_at,updated_at,deleted_at,resource_path,location_type,partner_internal_id,partner_internal_parent_id,parent_location_id,is_archived,access_path) VALUES
  ('01GWK39VWX66PNQGQQMEXT8MZ0','Center 5',now(),now(),NULL,'-2147483629','01GWK38KV1NZ7654N56B6WNVJ4','center_5','brand_1','01GWK39VVRSQDBFQY8XE6B21PN',false,'01GFMNHQ1WHGRC8AW6K913AM3G/01GWK39VVRSQDBFQY8XE6B21PN/01GWK39VWX66PNQGQQMEXT8MZ0'),
  ('01GWK39VWTEFF3DMH8MJCPARE0','Center 4',now(),now(),NULL,'-2147483629','01GWK38KV1NZ7654N56B6WNVJ4','center_4','brand_1','01GWK39VVRSQDBFQY8XE6B21PN',false,'01GFMNHQ1WHGRC8AW6K913AM3G/01GWK39VVRSQDBFQY8XE6B21PN/01GWK39VWTEFF3DMH8MJCPARE0'),
  ('01GWK39VWQH9PFMZET4WX9A1M7','Center 3',now(),now(),NULL,'-2147483629','01GWK38KV1NZ7654N56B6WNVJ4','center_3','brand_1','01GWK39VVRSQDBFQY8XE6B21PN',false,'01GFMNHQ1WHGRC8AW6K913AM3G/01GWK39VVRSQDBFQY8XE6B21PN/01GWK39VWQH9PFMZET4WX9A1M7'),
  ('01GWK39VWC6DZSB7DFGMWPBEGN','Center 2',now(),now(),NULL,'-2147483629','01GWK38KV1NZ7654N56B6WNVJ4','center_2','brand_1','01GWK39VVRSQDBFQY8XE6B21PN',false,'01GFMNHQ1WHGRC8AW6K913AM3G/01GWK39VVRSQDBFQY8XE6B21PN/01GWK39VWC6DZSB7DFGMWPBEGN'),
  ('01GWK39VVZ1FBCK4740RJQA0N8','Center 1',now(),now(),NULL,'-2147483629','01GWK38KV1NZ7654N56B6WNVJ4','center_1','brand_1','01GWK39VVRSQDBFQY8XE6B21PN',false,'01GFMNHQ1WHGRC8AW6K913AM3G/01GWK39VVRSQDBFQY8XE6B21PN/01GWK39VVZ1FBCK4740RJQA0N8'),
  ('01GWK39VVRSQDBFQY8XE6B21PN','Brand 1',now(),now(),NULL,'-2147483629','01GWK38KV0M6TM5869EH9VFEHC','brand_1',NULL,'01GFMNHQ1WHGRC8AW6K913AM3G',false,'01GFMNHQ1WHGRC8AW6K913AM3G/01GWK39VVRSQDBFQY8XE6B21PN')
  ON CONFLICT DO NOTHING;

INSERT INTO location_types (location_type_id,"name",display_name,parent_name,parent_location_type_id,updated_at,created_at,deleted_at,resource_path,is_archived,"level") VALUES
  ('01GWK438ZCQ81MT1GKRWRVGF8H','center','center','brand','01GWK438ZBYKESSM7NR6780Z4D',now(),now(),NULL,'-2147483630',false,2),
  ('01GWK438ZBYKESSM7NR6780Z4D','brand','brand','org','01GFMMFRXC6SKTTT44HWR3BRY8',now(),now(),NULL,'-2147483630',false,1)
  ON CONFLICT DO NOTHING;

INSERT INTO locations (location_id,"name",created_at,updated_at,deleted_at,resource_path,location_type,partner_internal_id,partner_internal_parent_id,parent_location_id,is_archived,access_path) VALUES
  ('01GWK5HJ1YAZEK3D9SPN08CT81','Center 5',now(),now(),NULL,'-2147483630','01GWK438ZCQ81MT1GKRWRVGF8H','center_5','brand_1','01GWK5HJ0NV171TDWQ9VWHH8HP',false,'01GFMMFRXC6SKTTT44HWR3BRY8/01GWK5HJ0NV171TDWQ9VWHH8HP/01GWK5HJ1YAZEK3D9SPN08CT81'),
  ('01GWK5HJ1W4KHK3RJSHXPR2F6T','Center 4',now(),now(),NULL,'-2147483630','01GWK438ZCQ81MT1GKRWRVGF8H','center_4','brand_1','01GWK5HJ0NV171TDWQ9VWHH8HP',false,'01GFMMFRXC6SKTTT44HWR3BRY8/01GWK5HJ0NV171TDWQ9VWHH8HP/01GWK5HJ1W4KHK3RJSHXPR2F6T'),
  ('01GWK5HJ1SRYDTEXXFH6R0DT1R','Center 3',now(),now(),NULL,'-2147483630','01GWK438ZCQ81MT1GKRWRVGF8H','center_3','brand_1','01GWK5HJ0NV171TDWQ9VWHH8HP',false,'01GFMMFRXC6SKTTT44HWR3BRY8/01GWK5HJ0NV171TDWQ9VWHH8HP/01GWK5HJ1SRYDTEXXFH6R0DT1R'),
  ('01GWK5HJ1B26A81W9GY36PE8WP','Center 2',now(),now(),NULL,'-2147483630','01GWK438ZCQ81MT1GKRWRVGF8H','center_2','brand_1','01GWK5HJ0NV171TDWQ9VWHH8HP',false,'01GFMMFRXC6SKTTT44HWR3BRY8/01GWK5HJ0NV171TDWQ9VWHH8HP/01GWK5HJ1B26A81W9GY36PE8WP'),
  ('01GWK5HJ0Y0382GATG246461QM','Center 1',now(),now(),NULL,'-2147483630','01GWK438ZCQ81MT1GKRWRVGF8H','center_1','brand_1','01GWK5HJ0NV171TDWQ9VWHH8HP',false,'01GFMMFRXC6SKTTT44HWR3BRY8/01GWK5HJ0NV171TDWQ9VWHH8HP/01GWK5HJ0Y0382GATG246461QM'),
  ('01GWK5HJ0NV171TDWQ9VWHH8HP','Brand 1',now(),now(),NULL,'-2147483630','01GWK438ZBYKESSM7NR6780Z4D','brand_1',NULL,'01GFMMFRXC6SKTTT44HWR3BRY8',false,'01GFMMFRXC6SKTTT44HWR3BRY8/01GWK5HJ0NV171TDWQ9VWHH8HP')
  ON CONFLICT DO NOTHING;

INSERT INTO courses (course_id,"name",country,subject,grade,display_order,updated_at,created_at,school_id,deleted_at,course_type,start_date,end_date,teacher_ids,preset_study_plan_id,icon,status,resource_path,teaching_method,course_type_id,remarks,is_archived,course_partner_id) VALUES
  ('01GWK5JB12FJC16STAMWWB0XT0','course 3',NULL,NULL,NULL,1,now(),now(),-2147483630,NULL,NULL,NULL,'2025-01-01 07:00:00+07',NULL,NULL,'','COURSE_STATUS_NONE','-2147483630','COURSE_TEACHING_METHOD_NONE',NULL,'',false,'1003'),
  ('01GWK5J5V1S3PDX2Z459X77YXR','course 2',NULL,NULL,NULL,1,now(),now(),-2147483630,NULL,NULL,NULL,'2025-01-01 07:00:00+07',NULL,NULL,'','COURSE_STATUS_NONE','-2147483630','COURSE_TEACHING_METHOD_NONE',NULL,'',false,'1002'),
  ('01GWK5J0HX6HVK8C5T2QSMJEKS','course 1',NULL,NULL,NULL,1,now(),now(),-2147483630,NULL,NULL,NULL,'2025-01-01 07:00:00+07',NULL,NULL,'','COURSE_STATUS_NONE','-2147483630','COURSE_TEACHING_METHOD_NONE',NULL,'',false,'1001'),
  ('01GWK3QZ99BJKDS4Q2MYJT95TS','course 3',NULL,NULL,NULL,1,now(),now(),-2147483629,NULL,NULL,NULL,'2025-01-01 07:00:00+07',NULL,NULL,'','COURSE_STATUS_NONE','-2147483629','COURSE_TEACHING_METHOD_NONE',NULL,'',false,'1003'),
  ('01GWK3QSBD4D0PKBXENQN1CPY6','course 2',NULL,NULL,NULL,1,now(),now(),-2147483629,NULL,NULL,NULL,'2025-01-01 07:00:00+07',NULL,NULL,'','COURSE_STATUS_NONE','-2147483629','COURSE_TEACHING_METHOD_NONE',NULL,'',false,'1002'),
  ('01GWK3NF2ZF2TKNGEN26YJQMPZ','course 1',NULL,NULL,NULL,1,now(),now(),-2147483629,NULL,NULL,NULL,'2025-01-01 07:00:00+07',NULL,NULL,'','COURSE_STATUS_NONE','-2147483629','COURSE_TEACHING_METHOD_NONE',NULL,'',false,'1001')
  ON CONFLICT DO NOTHING;

INSERT INTO course_access_paths (course_id,location_id,created_at,updated_at,deleted_at,resource_path) VALUES
	('01GWK3NF2ZF2TKNGEN26YJQMPZ','01GWK39VVZ1FBCK4740RJQA0N8',now(),now(),NULL,'-2147483629'),
	('01GWK3NF2ZF2TKNGEN26YJQMPZ','01GWK39VWC6DZSB7DFGMWPBEGN',now(),now(),NULL,'-2147483629'),
	('01GWK3NF2ZF2TKNGEN26YJQMPZ','01GWK39VWQH9PFMZET4WX9A1M7',now(),now(),NULL,'-2147483629'),
	('01GWK3NF2ZF2TKNGEN26YJQMPZ','01GWK39VWTEFF3DMH8MJCPARE0',now(),now(),NULL,'-2147483629'),
	('01GWK3NF2ZF2TKNGEN26YJQMPZ','01GWK39VWX66PNQGQQMEXT8MZ0',now(),now(),NULL,'-2147483629'),
	('01GWK3QSBD4D0PKBXENQN1CPY6','01GWK39VVZ1FBCK4740RJQA0N8',now(),now(),NULL,'-2147483629'),
	('01GWK3QSBD4D0PKBXENQN1CPY6','01GWK39VWC6DZSB7DFGMWPBEGN',now(),now(),NULL,'-2147483629'),
	('01GWK3QSBD4D0PKBXENQN1CPY6','01GWK39VWQH9PFMZET4WX9A1M7',now(),now(),NULL,'-2147483629'),
	('01GWK3QSBD4D0PKBXENQN1CPY6','01GWK39VWTEFF3DMH8MJCPARE0',now(),now(),NULL,'-2147483629'),
	('01GWK3QSBD4D0PKBXENQN1CPY6','01GWK39VWX66PNQGQQMEXT8MZ0',now(),now(),NULL,'-2147483629'),
	('01GWK3QZ99BJKDS4Q2MYJT95TS','01GWK39VVZ1FBCK4740RJQA0N8',now(),now(),NULL,'-2147483629'),
	('01GWK3QZ99BJKDS4Q2MYJT95TS','01GWK39VWC6DZSB7DFGMWPBEGN',now(),now(),NULL,'-2147483629'),
	('01GWK3QZ99BJKDS4Q2MYJT95TS','01GWK39VWQH9PFMZET4WX9A1M7',now(),now(),NULL,'-2147483629'),
	('01GWK3QZ99BJKDS4Q2MYJT95TS','01GWK39VWTEFF3DMH8MJCPARE0',now(),now(),NULL,'-2147483629'),
	('01GWK3QZ99BJKDS4Q2MYJT95TS','01GWK39VWX66PNQGQQMEXT8MZ0',now(),now(),NULL,'-2147483629'),
	('01GWK5J0HX6HVK8C5T2QSMJEKS','01GWK5HJ0Y0382GATG246461QM',now(),now(),NULL,'-2147483630'),
	('01GWK5J0HX6HVK8C5T2QSMJEKS','01GWK5HJ1B26A81W9GY36PE8WP',now(),now(),NULL,'-2147483630'),
	('01GWK5J0HX6HVK8C5T2QSMJEKS','01GWK5HJ1SRYDTEXXFH6R0DT1R',now(),now(),NULL,'-2147483630'),
	('01GWK5J0HX6HVK8C5T2QSMJEKS','01GWK5HJ1W4KHK3RJSHXPR2F6T',now(),now(),NULL,'-2147483630'),
	('01GWK5J0HX6HVK8C5T2QSMJEKS','01GWK5HJ1YAZEK3D9SPN08CT81',now(),now(),NULL,'-2147483630'),
	('01GWK5J5V1S3PDX2Z459X77YXR','01GWK5HJ0Y0382GATG246461QM',now(),now(),NULL,'-2147483630'),
	('01GWK5J5V1S3PDX2Z459X77YXR','01GWK5HJ1B26A81W9GY36PE8WP',now(),now(),NULL,'-2147483630'),
	('01GWK5J5V1S3PDX2Z459X77YXR','01GWK5HJ1SRYDTEXXFH6R0DT1R',now(),now(),NULL,'-2147483630'),
	('01GWK5J5V1S3PDX2Z459X77YXR','01GWK5HJ1W4KHK3RJSHXPR2F6T',now(),now(),NULL,'-2147483630'),
	('01GWK5J5V1S3PDX2Z459X77YXR','01GWK5HJ1YAZEK3D9SPN08CT81',now(),now(),NULL,'-2147483630'),
	('01GWK5JB12FJC16STAMWWB0XT0','01GWK5HJ0Y0382GATG246461QM',now(),now(),NULL,'-2147483630'),
	('01GWK5JB12FJC16STAMWWB0XT0','01GWK5HJ1B26A81W9GY36PE8WP',now(),now(),NULL,'-2147483630'),
	('01GWK5JB12FJC16STAMWWB0XT0','01GWK5HJ1SRYDTEXXFH6R0DT1R',now(),now(),NULL,'-2147483630'),
	('01GWK5JB12FJC16STAMWWB0XT0','01GWK5HJ1W4KHK3RJSHXPR2F6T',now(),now(),NULL,'-2147483630'),
	('01GWK5JB12FJC16STAMWWB0XT0','01GWK5HJ1YAZEK3D9SPN08CT81',now(),now(),NULL,'-2147483630')
  ON CONFLICT DO NOTHING;

-- create user group with role teacher locally
INSERT INTO public.user_group
  (user_group_id, user_group_name, is_system, created_at, updated_at, resource_path)
VALUES
  ('01GCZX3N1HFVRQ3XNZT547B8EB', 'Teacher', true, now(), now(), '-2147483648'),
  ('01GCZX3N1HFVRQ3XNZT547B8EG', 'Teacher', true, now(), now(), '-2147483647')
  ON CONFLICT DO NOTHING;

INSERT INTO public.granted_role
  (granted_role_id, user_group_id, role_id, created_at, updated_at, resource_path)
VALUES
  ('01GCZX3N1Q5RE5ECY1KCQ697V0', '01GCZX3N1HFVRQ3XNZT547B8EB', '01G1GQEKEHXKSM78NBW96NJ7H0', now(), now(), '-2147483648'),
  ('01GCZX3N1Q5RE5ECY1KCQ697V9', '01GCZX3N1HFVRQ3XNZT547B8EG', '01G1GQEKEHXKSM78NBW96NJ7H2', now(), now(), '-2147483647')
  ON CONFLICT DO NOTHING;

INSERT INTO public.granted_role_access_path
  (granted_role_id, location_id, created_at, updated_at, resource_path)
VALUES
  ('01GCZX3N1Q5RE5ECY1KCQ697V0', '01FR4M51XJY9E77GSN4QZ1Q9N1', now(), now(), '-2147483648'),
  ('01GCZX3N1Q5RE5ECY1KCQ697V9', '01FR4M51XJY9E77GSN4QZ1Q9N2', now(), now(), '-2147483647')
  ON CONFLICT DO NOTHING;

-- prevent flaky tests when kafka can not sync data
INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) 
VALUES
('0b6f273d-d249-4b31-8d4d-c69f51721b9d', 'user.enrollment.update_status_manual', 'on', 'string', NULL, '2022-11-09 17:48:57.663', '2022-11-09 17:48:57.663', NULL, '-2147483644'),
('0b6f273d-d249-4b31-8d4d-c69f51721b35', 'user.enrollment.update_status_manual', 'off', 'string', NULL, '2022-11-09 17:48:57.663', '2022-11-09 17:48:57.663', NULL, '-2147483635'),
('0b6f273d-d249-4b31-8d4d-c69f51721b48', 'user.enrollment.update_status_manual', 'on', 'string', NULL, '2022-11-09 17:48:57.663', '2022-11-09 17:48:57.663', NULL, '-2147483648'),
('0b6f273d-d249-4b31-8d4d-c69f51721b30', 'user.enrollment.update_status_manual', 'on', 'string', NULL, '2022-11-09 17:48:57.663', '2022-11-09 17:48:57.663', NULL, '-2147483630'),
('0b6f273d-d249-4b31-8d4d-c69f51721b29', 'user.enrollment.update_status_manual', 'on', 'string', NULL, '2022-11-09 17:48:57.663', '2022-11-09 17:48:57.663', NULL, '-2147483629')
ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;

INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) 
VALUES
('9e2052ce-10fe-46b8-8568-be119457971d', 'user.student_management.deactivate_parent', 'on', 'string', NULL, now(), now(), NULL, '-2147483644'),
('ef049623-1b7c-47de-9ed0-63c9883020e3', 'user.student_management.deactivate_parent', 'on', 'string', NULL, now(), now(), NULL, '-2147483648'),
('40f6fcdb-11d0-42ea-8682-5cd82591d445', 'user.student_management.deactivate_parent', 'on', 'string', NULL, now(), now(), NULL, '-2147483630'),
('196b385b-db80-4d91-af1a-f17965931e18', 'user.student_management.deactivate_parent', 'on', 'string', NULL, now(), now(), NULL, '-2147483629')
ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO UPDATE SET config_value = 'on';

-- prevent flaky tests when kafka can not sync data
INSERT INTO public.user_access_paths
(user_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('user_id_existing_01', '-2147483648_location-id-1', now(), now(), NULL, '-2147483648'),
    ('user_id_existing_02', '-2147483648_location-id-2', now(), now(), NULL, '-2147483648'),
    ('user_id_existing_03', '-2147483648_location-id-3', now(), now(), NULL, '-2147483648'),
    ('user_id_existing_04', '-2147483648_location-id-4', now(), now(), NULL, '-2147483648'),
    ('user_id_existing_05', '-2147483648_location-id-4', now(), now(), NULL, '-2147483648'),
    ('01GK1ETTR4ECP5DSAKFMAM24KV003', '-2147483648_location-id-2', now(), now(), NULL, '-2147483648')
ON CONFLICT DO NOTHING;

INSERT INTO public.internal_configuration_value (configuration_id, config_key, config_value, config_value_type, last_editor, created_at, updated_at, deleted_at, resource_path) 
VALUES
('95dd703f-e7cb-4fd4-9b59-2b44e598946e', 'user.auth.username', 'on', 'string', NULL, now(), now(), NULL, '-2147483644'),
('b62595ee-6107-4ea5-bcf4-b19fd68efaf8', 'user.auth.username', 'on', 'string', NULL, now(), now(), NULL, '-2147483648'),
('01bb3183-bb2e-4959-a7a7-eccc738a6f05', 'user.auth.username', 'on', 'string', NULL, now(), now(), NULL, '-2147483630'),
('4fc12054-6b60-4abc-b6ea-8772468a5779', 'user.auth.username', 'on', 'string', NULL, now(), now(), NULL, '-2147483629'),
('d073fe4d-d3db-45d2-8490-0057f8a09319', 'user.auth.username', 'on', 'string', NULL, now(), now(), NULL, '-2147483635')
ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO UPDATE SET config_value = 'on';

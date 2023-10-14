\connect bob;
INSERT INTO public.students(student_id, current_grade, billing_date, updated_at, created_at, resource_path) VALUES
    ('01GK1ETTR4ECP5DSAKFMAM24KV', 1, now(), now(), now(), '-2147483648');
TRUNCATE TABLE public.students;

INSERT INTO public.lessons (lesson_id,teacher_id,course_id,created_at,updated_at,deleted_at,end_at,control_settings,lesson_group_id,room_id,lesson_type,status,stream_learner_counter,learner_ids,"name",start_time,end_time,resource_path,room_state,teaching_model,class_id,center_id,teaching_method,teaching_medium,scheduling_status,is_locked,scheduler_id) VALUES
    ('01G5ZJPD19583SRNTASC5B6HMB',
     '01G5ZJPD19583SRNTASC5B6HMK',
     '01G5ZJPD19583SRNTASC5B6HMK',
     '2022-08-03 15:47:37.629',
     '2022-08-03 15:47:37.629',NULL,NULL,NULL,NULL,NULL,NULL,NULL,0,'{}','test',NULL,NULL,'16091',NULL,NULL,NULL,'01FR4M51XJY9E77GSN4QZ1Q911',NULL,NULL,'LESSON_SCHEDULING_STATUS_PUBLISHED',false,NULL);

TRUNCATE TABLE public.lessons;
INSERT INTO public.lesson_members (lesson_id,user_id,updated_at,created_at,deleted_at,resource_path,attendance_status,attendance_remark,course_id,attendance_notice,attendance_reason,attendance_note,user_first_name,user_last_name) VALUES
    ('01G5ZJPD19583SRNTASC5B6HMB','01GGP4SS7CNK30HXKC9CMKAZB2','2023-02-02 10:10:56.514','2023-02-02 10:10:56.514',NULL,'16091',NULL,NULL,NULL,NULL,NULL,NULL,NULL,NULL);

TRUNCATE TABLE public.lessons;

INSERT INTO public.users
(user_id, user_external_id, username, login_email, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system, user_role)
VALUES
('01GK1ETTR4ECP5DSAKFMAM24KV001','user_external_id_01GK1ETTR4ECP5DSAKFMAM24KV001','username_jprep','username_jprep', 'COUNTRY_JP', 'Temporary Student', '', NULL, 'temporary_student+usermgmt001@manabie.com', NULL, NULL, 'USER_GROUP_STUDENT', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483642', true, 'student'),
('user_id_existing_01', 'user_external_id_existing_01', 'username_existing_01@gmail.com', 'username_existing_01@gmail.com', 'COUNTRY_JP', '[Manabie] Student Existing 01', '', NULL, 'existing_email_01@manabie.com', NULL, NULL, 'USER_GROUP_STUDENT', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483648', true, 'student'),
('user_id_existing_02', 'user_external_id_existing_02', 'username_existing_02@gmail.com', 'username_existing_02@gmail.com', 'COUNTRY_JP', '[Manabie] Student Existing 02', '', NULL, 'existing_email_02@manabie.com', NULL, NULL, 'USER_GROUP_STUDENT', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483648', true, 'student'),
('user_id_existing_03', 'user_external_id_existing_03', 'username_existing_03@gmail.com', 'username_existing_03@gmail.com', 'COUNTRY_JP', '[Manabie] Student Existing 03', '', NULL, 'existing_email_03@manabie.com', NULL, NULL, 'USER_GROUP_STUDENT', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483648', true, 'student'),
('user_id_existing_04', 'user_external_id_existing_04', 'username_existing_04@gmail.com', 'username_existing_04@gmail.com', 'COUNTRY_JP', '[Manabie] Parent Existing 03', '', NULL, 'existing_email_04@manabie.com', NULL, NULL, 'USER_GROUP_PARENT', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483648', true, 'parent'),
('user_id_existing_05', 'user_external_id_existing_05', 'existing_email_05@manabie.com', 'existing_email_05@manabie.com', 'COUNTRY_JP', '[Manabie] Student Existing 05', '', NULL, 'existing_email_05@manabie.com', NULL, NULL, 'USER_GROUP_STUDENT', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483648', true, 'student')
    ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES
('user_id_existing_01', '-2147483648_location-id-1', now(), now(), NULL, '-2147483648'),
('user_id_existing_02', '-2147483648_location-id-2', now(), now(), NULL, '-2147483648'),
('user_id_existing_03', '-2147483648_location-id-3', now(), now(), NULL, '-2147483648'),
('user_id_existing_04', '-2147483648_location-id-1', now(), now(), NULL, '-2147483648'),
('user_id_existing_05', '-2147483648_location-id-4', now(), now(), NULL, '-2147483648')
    ON CONFLICT DO NOTHING;

INSERT INTO public.students(student_id, current_grade, billing_date, updated_at, created_at, resource_path) VALUES
    ('01GK1ETTR4ECP5DSAKFMAM24KV001', 1, now(), now(), now(), '-2147483642'),
    ('user_id_existing_01', 1, now(), now(), now(), '-2147483648'),
    ('user_id_existing_02', 1, now(), now(), now(), '-2147483648'),
    ('user_id_existing_03', 1, now(), now(), now(), '-2147483648'),
    ('user_id_existing_05', 1, now(), now(), now(), '-2147483648')
    ON CONFLICT DO NOTHING;

INSERT INTO public.users
(user_external_id,user_id, country, name, avatar, phone_number, email, device_token, allow_notification, user_group, updated_at, created_at, is_tester, facebook_id, platform, phone_verified, email_verified, deleted_at, given_name, resource_path, is_system)
VALUES
    ('parent_external_id_existing_01', '01GK1ETTR4ECP5DSAKFMAM24KV002', 'COUNTRY_JP', 'Temporary Parent 1', '', NULL, 'temporary_parent+usermgmt001@manabie.com', NULL, NULL, 'USER_GROUP_STUDENT', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483642', true),
    ('parent_external_id_existing_02', '01GK1ETTR4ECP5DSAKFMAM24KV003', 'COUNTRY_JP', 'Temporary Parent 2', '', NULL, 'temporary_parent+usermgmt001@manabie.com', NULL, NULL, 'USER_GROUP_STUDENT', now(), now(), NULL, NULL, NULL, NULL, NULL, NULL, NULL, '-2147483648', true)
    ON CONFLICT DO NOTHING;

INSERT INTO public.parents
(parent_id, created_at, updated_at, resource_path)
VALUES
('01GK1ETTR4ECP5DSAKFMAM24KV002', now(), now(), '-2147483642'),
('01GK1ETTR4ECP5DSAKFMAM24KV003', now(), now(), '-2147483648'),
('user_id_existing_04', now(), now(), '-2147483648')
    ON CONFLICT DO NOTHING;

INSERT INTO public.student_parents 
(student_id, parent_id, created_at, updated_at, deleted_at, resource_path, relationship)
VALUES
    ('01GK1ETTR4ECP5DSAKFMAM24KV001', '01GK1ETTR4ECP5DSAKFMAM24KV002', now(), now(), NULL, '-2147483642', ''),
    ('user_id_existing_02', '01GK1ETTR4ECP5DSAKFMAM24KV002', now(), now(), NULL, '-2147483648', ''),
    ('user_id_existing_01', 'user_id_existing_04', now(), now(), NULL, '-2147483648', '')
ON CONFLICT DO NOTHING;

INSERT INTO public.user_access_paths
(user_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('01GK1ETTR4ECP5DSAKFMAM24KV003', '-2147483648_location-id-2', now(), now(), NULL, '-2147483648')
ON CONFLICT DO NOTHING;

INSERT INTO public.student_enrollment_status_history 
(student_id, location_id, enrollment_status, start_date, created_at, updated_at, deleted_at, resource_path)
VALUES ('01GK1ETTR4ECP5DSAKFMAM24KV001', '01FR4M51XJY9E77GSN4QZ1Q9N7', 'STUDENT_ENROLLMENT_STATUS_ENROLLED', now(), now(), now(), NULL, '-2147483642')
ON CONFLICT DO NOTHING;

-- DWH KEC
INSERT INTO public.grade
("name", is_archived, updated_at, created_at, resource_path, grade_id, partner_internal_id, deleted_at, "sequence")
VALUES('高校3年生', false, '2022-12-24 04:39:11.543', '2022-08-24 15:08:41.738', '-2147483642', '01GB7EAYGAS312J0J2W7H0JR1242', 'id_12', NULL, 12)
ON CONFLICT ON CONSTRAINT grade_pk DO UPDATE SET resource_path = '-2147483642';

INSERT INTO public.grade
("name", is_archived, updated_at, created_at, resource_path, grade_id, partner_internal_id, deleted_at, "sequence")
VALUES('小学4年生', false, '2022-12-24 04:39:11.543', '2022-08-24 15:08:41.738', '-2147483642', '01GB7EAYGAS312J0J2W87JR4N242', 'id_4', NULL, 4)
ON CONFLICT ON CONSTRAINT grade_pk DO UPDATE SET resource_path = '-2147483642';

INSERT INTO public.grade
("name", is_archived, updated_at, created_at, resource_path, grade_id, partner_internal_id, deleted_at, "sequence")
VALUES('小学3年生', false, '2022-12-24 04:39:11.543', '2022-08-24 15:08:41.738', '-2147483642', '01GB7EAYGAS312J0J2W5APGQYX42', 'id_3', NULL, 3)
ON CONFLICT ON CONSTRAINT grade_pk DO UPDATE SET resource_path = '-2147483642';

INSERT INTO public.grade
("name", is_archived, updated_at, created_at, resource_path, grade_id, partner_internal_id, deleted_at, "sequence")
VALUES('中学1年生', false, '2022-12-24 04:39:11.543', '2022-08-24 15:08:41.738', '-2147483642', '01GB7EAYGAS312J0J2VT50BC6742', 'id_7', NULL, 7)
ON CONFLICT ON CONSTRAINT grade_pk DO UPDATE SET resource_path = '-2147483642';

INSERT INTO public.grade
("name", is_archived, updated_at, created_at, resource_path, grade_id, partner_internal_id, deleted_at, "sequence")
VALUES('中学2年生', false, '2022-12-24 04:39:11.543', '2022-08-24 15:08:41.738', '-2147483642', '01GB7EAYGAS312J0J2W5CRR9RH42', 'id_8', NULL, 8)
ON CONFLICT ON CONSTRAINT grade_pk DO UPDATE SET resource_path = '-2147483642';

INSERT INTO public.grade
("name", is_archived, updated_at, created_at, resource_path, grade_id, partner_internal_id, deleted_at, "sequence")
VALUES('中学3年生', false, '2022-12-24 04:39:11.543', '2022-08-24 15:08:41.738', '-2147483642', '01GB7EAYGAS312J0J2VV98FJSB42', 'id_9', NULL, 9)
ON CONFLICT ON CONSTRAINT grade_pk DO UPDATE SET resource_path = '-2147483642';

INSERT INTO public.school_level
(school_level_id, school_level_name, "sequence", is_archived, created_at, updated_at, deleted_at, resource_path)
VALUES('01G9KKDJA8QCDM9BWFNM5QTWSB42', '[Aug 18] School Level 2', 2, false, '2022-08-04 11:56:59.976', '2022-08-18 11:27:04.619', NULL, '-2147483642') ON CONFLICT DO NOTHING;
INSERT INTO public.school_level
(school_level_id, school_level_name, "sequence", is_archived, created_at, updated_at, deleted_at, resource_path)
VALUES('01G9KKDJA9AP1HAB1SRN6RYAZK42', '[Aug 18] School Level 3', 3, false, '2022-08-04 11:56:59.977', '2022-08-18 11:27:04.619', NULL, '-2147483642') ON CONFLICT DO NOTHING;
INSERT INTO public.school_level
(school_level_id, school_level_name, "sequence", is_archived, created_at, updated_at, deleted_at, resource_path)
VALUES('01GA5EPWQRHW0FK2VWKP4Z7FY342', '[Aug 18] School Level 4', 4, false, '2022-08-11 10:21:02.839', '2022-08-18 11:27:04.619', NULL, '-2147483642') ON CONFLICT DO NOTHING;
INSERT INTO public.school_level
(school_level_id, school_level_name, "sequence", is_archived, created_at, updated_at, deleted_at, resource_path)
VALUES('01GAQK8TZTZKC6TPBDD1DAYJAM42', '[Aug 18] School Level 5', 5, false, '2022-08-18 11:27:04.619', '2022-08-18 11:27:04.619', NULL, '-2147483642') ON CONFLICT DO NOTHING;

INSERT INTO public.school_level_grade
(school_level_id, grade_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01G9KKDJA8QCDM9BWFNM5QTWSB42', '01GB7EAYGAS312J0J2W7H0JR1242', '2022-08-04 11:56:59.976', '2022-08-18 11:27:04.619', NULL, '-2147483642') ON CONFLICT DO NOTHING;
INSERT INTO public.school_level_grade
(school_level_id, grade_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01G9KKDJA9AP1HAB1SRN6RYAZK42', '01GB7EAYGAS312J0J2W87JR4N242', '2022-08-04 11:56:59.977', '2022-08-18 11:27:04.619', NULL, '-2147483642') ON CONFLICT DO NOTHING;
INSERT INTO public.school_level_grade
(school_level_id, grade_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GA5EPWQRHW0FK2VWKP4Z7FY342', '01GB7EAYGAS312J0J2W5APGQYX42', '2022-08-11 10:21:02.839', '2022-08-18 11:27:04.619', NULL, '-2147483642') ON CONFLICT DO NOTHING;
INSERT INTO public.school_level_grade
(school_level_id, grade_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GAQK8TZTZKC6TPBDD1DAYJAM42', '01GB7EAYGAS312J0J2VT50BC6742', '2022-08-18 11:27:04.619', '2022-08-18 11:27:04.619', NULL, '-2147483642') ON CONFLICT DO NOTHING;

INSERT INTO public.school_info
(school_id, school_name, school_name_phonetic, is_archived, created_at, updated_at, deleted_at, resource_path, school_level_id, address, school_partner_id)
VALUES('01G9M1QKTC8N1XQY7GA11QYMAG42', 'School Name 2', 'School Name Phonetic 2', false, '2022-08-04 16:07:09.260', '2022-08-04 16:07:09.260', NULL, '-2147483642', '01G9KKDJA9AP1HAB1SRN6RYAZK42', '124 Nguyen Van Linh, Phuong 4, Quan 2, Thanh pho Ho Chi Minh', 'school_2') ON CONFLICT DO NOTHING;

INSERT INTO public.school_info
(school_id, school_name, school_name_phonetic, is_archived, created_at, updated_at, deleted_at, resource_path, school_level_id, address, school_partner_id)
VALUES('01G9M1QKTC8N1XQY7GA11QYMAJ42', 'School Name 4', 'School Name Phonetic 4', false, '2022-08-04 16:07:09.260', '2022-08-04 16:07:09.260', NULL, '-2147483642', '01G9KKDJA8QCDM9BWFNM5QTWSB42', '227 Nguyen Van Cu, Phuong 4, Quan 5, Thanh pho Ho Chi Minh', 'school_4') ON CONFLICT DO NOTHING;

INSERT INTO public.school_info
(school_id, school_name, school_name_phonetic, is_archived, created_at, updated_at, deleted_at, resource_path, school_level_id, address, school_partner_id)
VALUES('01GA5JPW36VSGZYXKB1NCVH5YZ42', 'School 1', 'School 1', false, '2022-08-11 11:30:56.209', '2022-08-11 11:40:07.709', NULL, '-2147483642', '01GA5EPWQRHW0FK2VWKP4Z7FY342', '', 'school-1') ON CONFLICT DO NOTHING;   

INSERT INTO public.school_course
(school_course_id, school_course_name, school_course_name_phonetic, school_id, is_archived, created_at, updated_at, deleted_at, resource_path, school_course_partner_id)
VALUES('01G9RPYBFX6N81WB2K273916HT42', 'English advanced speaking', 'English advanced speaking', '01G9M1QKTC8N1XQY7GA11QYMAG42', false, '2022-08-06 11:34:47.933', '2022-08-06 11:34:47.933', NULL, '-2147483642', 'school_course_2') ON CONFLICT DO NOTHING;

INSERT INTO public.school_course
(school_course_id, school_course_name, school_course_name_phonetic, school_id, is_archived, created_at, updated_at, deleted_at, resource_path, school_course_partner_id)
VALUES('01GA5N7DP9639GSTG2NVR341WT42', 'School Course 3', 'School Course 3', '01GA5JPW36VSGZYXKB1NCVH5YZ42', false, '2022-08-11 12:14:55.600', '2022-08-11 12:16:57.850', NULL, '-2147483642', 'school_course_3') ON CONFLICT DO NOTHING;

INSERT INTO public.school_course
(school_course_id, school_course_name, school_course_name_phonetic, school_id, is_archived, created_at, updated_at, deleted_at, resource_path, school_course_partner_id)
VALUES('01GB7W3WT9AD8SQ9BTSJGQT16E42', 'E2E - School Course 1661342969247-pakA9', 'E2E - School Course 1661342969247-pakA9', '01G9M1QKTC8N1XQY7GA11QYMAG42', false, '2022-08-24 19:09:30.728', '2022-08-24 19:09:30.728', NULL, '-2147483642', 'e2e-school-course-1661342969247-pakA9') ON CONFLICT DO NOTHING;

INSERT INTO public.school_course
(school_course_id, school_course_name, school_course_name_phonetic, school_id, is_archived, created_at, updated_at, deleted_at, resource_path, school_course_partner_id)
VALUES('01GB7W3WT9AD8SQ9BTSJGQT16F42', 'E2E - School Course 1661342969247-5ZGYC', 'E2E - School Course 1661342969247-5ZGYC', '01GA5JPW36VSGZYXKB1NCVH5YZ42', false, '2022-08-24 19:09:30.728', '2022-08-24 19:09:30.728', NULL, '-2147483642', 'e2e-school-course-1661342969247-5ZGYC') ON CONFLICT DO NOTHING;

INSERT INTO public.school_course
(school_course_id, school_course_name, school_course_name_phonetic, school_id, is_archived, created_at, updated_at, deleted_at, resource_path, school_course_partner_id)
VALUES('01GB7W3WT9AD8SQ9BTSJGQT16H42', 'E2E - School Course 1661342969247-eerJ0', 'E2E - School Course 1661342969247-eerJ0', '01G9M1QKTC8N1XQY7GA11QYMAJ42', false, '2022-08-24 19:09:30.728', '2022-08-24 19:09:30.728', NULL, '-2147483642', 'e2e-school-course-1661342969247-eerJ0') ON CONFLICT DO NOTHING;

INSERT INTO public.school_history
(student_id, school_id, school_course_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GK1ETTR4ECP5DSAKFMAM24KV001', '01G9M1QKTC8N1XQY7GA11QYMAJ42', '01GB7W3WT9AD8SQ9BTSJGQT16H42', '2022-08-24 19:09:30.728', '2022-08-24 19:09:30.728', NULL, '-2147483642') ON CONFLICT DO NOTHING;

-- school_level, school_info, school_course, school_grade_level
INSERT INTO public.school_level
	(school_level_id, school_level_name, "sequence", is_archived, created_at, updated_at, deleted_at, resource_path)
VALUES
    ('-2147483648_school_level_01', '[Manabie] School Level 1', 10001, false, now(), now(), NULL, '-2147483648'),
    ('-2147483648_school_level_02', '[Manabie] School Level 2', 10002, false, now(), now(), NULL, '-2147483648'),
    ('-2147483648_school_level_03', '[Manabie] School Level 3', 10003, false, now(), now(), NULL, '-2147483648'),
    ('-2147483635_school_level_01', '[KEC-Demo] School Level 1', 10004, false, now(), now(), NULL, '-2147483635'),
    ('-2147483635_school_level_02', '[KEC-Demo] School Level 2', 10005, false, now(), now(), NULL, '-2147483635'),
    ('-2147483635_school_level_03', '[KEC-Demo] School Level 3', 10006, false, now(), now(), NULL, '-2147483635') 
ON CONFLICT DO NOTHING;

INSERT INTO public.school_info
	(school_id, school_name, school_name_phonetic, is_archived, created_at, updated_at, deleted_at, resource_path, school_level_id, address, school_partner_id)
VALUES
    ('-2147483648_school_id_01', '[Manabie] School Name 1', '[Manabie] School Name Phonetic 1', false, now(), now(), NULL, '-2147483648', '-2147483648_school_level_01', '124 Nguyen Van Linh, Phuong 4, Quan 2, Thanh pho Ho Chi Minh', 'school_partner_id_1'),
    ('-2147483648_school_id_02', '[Manabie] School Name 2', '[Manabie] School Name Phonetic 2', false, now(), now(), NULL, '-2147483648', '-2147483648_school_level_02', '12 Ton Dan, Phuong 13, Quan 4, Thanh pho Ho Chi Minh', 'school_partner_id_2'),
    ('-2147483648_school_id_03', '[Manabie] School Name 3', '[Manabie] School Name Phonetic 3', false, now(), now(), NULL, '-2147483648', '-2147483648_school_level_03', '001 Nguyen Gia Tri, Phuong 13, Quan Binh Thanh, Thanh pho Ho Chi Minh', 'school_partner_id_3'),
    ('-2147483635_school_id_01', '[KEC-Demo] School Name 1', '[KEC-Demo] School Name Phonetic 1', false, now(), now(), NULL, '-2147483635', '-2147483635_school_level_01', '147 Ton Dat Tien, Phuong Tan Phong, Quan 7, Thanh pho Ho Chi Minh', 'school_partner_id_1'),
    ('-2147483635_school_id_02', '[KEC-Demo] School Name 2', '[KEC-Demo] School Name Phonetic 2', false, now(), now(), NULL, '-2147483635', '-2147483635_school_level_02', '1 Le Duan, Phuong Ben Nghe, Quan 1, Thanh pho Ho Chi Minh', 'school_partner_id_2'),
    ('-2147483635_school_id_03', '[KEC-Demo] School Name 3', '[KEC-Demo] School Name Phonetic 3', false, now(), now(), NULL, '-2147483635', '-2147483635_school_level_03', '001 Nguyen Gia Tri, Phuong 13, Quan Binh Thanh, Thanh pho Ho Chi Minh', 'school_partner_id_3')
    ON CONFLICT DO NOTHING;

INSERT INTO public.school_course
	(school_course_id, school_course_name, school_course_name_phonetic, school_id, is_archived, created_at, updated_at, deleted_at, resource_path, school_course_partner_id)
VALUES
    ('-2147483648_school_course_id_01', '[Manabie] School Course 1', '[Manabie] School Course Phonetic 1', '-2147483648_school_id_01', false, now(), now(), NULL, '-2147483648', 'school_course_partner_id_01'),
    ('-2147483648_school_course_id_02', '[Manabie] School Course 2', '[Manabie] School Course Phonetic 2', '-2147483648_school_id_02', false, now(), now(), NULL, '-2147483648', 'school_course_partner_id_02'),
    ('-2147483648_school_course_id_03', '[Manabie] School Course 3', '[Manabie] School Course Phonetic 3', '-2147483648_school_id_03', false, now(), now(), NULL, '-2147483648', 'school_course_partner_id_03'),
    ('-2147483635_school_course_id_01', '[KEC-Demo] School Course 1', '[KEC-Demo] School Course Phonetic 1', '-2147483635_school_id_01', false, now(), now(), NULL, '-2147483635', 'school_course_partner_id_01'),
    ('-2147483635_school_course_id_02', '[KEC-Demo] School Course 2', '[KEC-Demo] School Course Phonetic 2', '-2147483635_school_id_02', false, now(), now(), NULL, '-2147483635', 'school_course_partner_id_02'),
    ('-2147483635_school_course_id_03', '[KEC-Demo] School Course 3', '[KEC-Demo] School Course Phonetic 3', '-2147483635_school_id_03', false, now(), now(), NULL, '-2147483635', 'school_course_partner_id_03') 
    ON CONFLICT DO NOTHING;
 -- current school
INSERT INTO public.school_level_grade
    (school_level_id, grade_id, created_at, updated_at, deleted_at, resource_path)
VALUES 
    ('-2147483648_school_level_01', '-2147483648_grade_01', now(), now(), NULL, '-2147483648')
ON CONFLICT DO NOTHING;

INSERT INTO public.user_tag
(user_tag_id, user_tag_name, user_tag_type, is_archived, created_at, updated_at, deleted_at, resource_path, user_tag_partner_id)
VALUES('01GAQMY5T223FDNWN7TJF01H8742', 'Tag Student 1', 'USER_TAG_TYPE_STUDENT', false, '2022-08-18 11:56:12.357', '2022-08-18 11:56:12.357', NULL, '-2147483642', 'tag-student-1')
ON CONFLICT DO NOTHING;

INSERT INTO public.tagged_user
(user_id, tag_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GK1ETTR4ECP5DSAKFMAM24KV001', '01GAQMY5T223FDNWN7TJF01H8742', '2023-03-11 16:22:25.918', '2023-03-11 16:22:25.918', NULL, '-2147483642')
ON CONFLICT DO NOTHING;

INSERT INTO public.user_phone_number
(user_phone_number_id, user_id, phone_number, "type", updated_at, created_at, deleted_at, resource_path)
VALUES('01GBS88KRQC35KV47W5FSNZE3B42', '01GK1ETTR4ECP5DSAKFMAM24KV001', '810312345678', 'STUDENT_PHONE_NUMBER', '2022-08-31 13:08:53.527', '2022-08-31 13:08:53.527', '2022-09-02 14:52:06.389', '-2147483642')
ON CONFLICT DO NOTHING;

INSERT INTO public.user_address
(user_address_id, user_id, address_type, postal_code, prefecture_id, city, created_at, updated_at, deleted_at, resource_path, first_street, second_street)
VALUES('01GV7ZMXW5E2TZZ9C1RJ9FEPB642', '01GK1ETTR4ECP5DSAKFMAM24KV001', 'HOME_ADDRESS', '542687', '01G94NBXEY8QX435T3JM3GK7Z9', '福島市', '2023-03-11 16:22:25.826', '2023-03-11 16:22:25.826', NULL, '-2147483642', '青柳町', '1-2')
ON CONFLICT DO NOTHING;

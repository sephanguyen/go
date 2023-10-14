\connect bob;
INSERT INTO public.locations
(location_id, "name", created_at, updated_at, deleted_at, resource_path, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, is_archived, access_path)
VALUES('01GV032YZ8FA4JGEAR4XXQX6L4', 'Location DWH test', timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483642', NULL, '', '', NULL, false, '01GV032YZ8FA4JGEAR4XXQX6L3');

INSERT INTO public.location_types
(location_type_id, "name", display_name, parent_name, parent_location_type_id, updated_at, created_at, deleted_at, resource_path, is_archived, "level")
VALUES('01GV032YZ8FA4JGEAR4XXQX6L5', 'Location type DWH test', 'brand', 'org', NULL, timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483642', false, 0);


INSERT INTO public.subject
(subject_id, "name", created_at, updated_at, deleted_at, resource_path)
VALUES('01GV032YZ8FA4JGEAR4XXQX3L6', 'Subject DWH test', timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483642');

INSERT INTO public.course_type
(course_type_id, "name", created_at, updated_at, deleted_at, resource_path, remarks, is_archived)
VALUES('01GWK5JB12FJC16STACOURSETYPE', 'course type DWH test', '2023-07-19 17:49:25.803', '2023-07-19 17:49:25.803', NULL, '-2147483642', '', false);

INSERT INTO public.courses
(course_id, "name", country, subject, grade, display_order, updated_at, created_at, school_id, deleted_at, course_type, start_date, end_date, teacher_ids, preset_study_plan_id, icon, status, resource_path, teaching_method, course_type_id, remarks, is_archived, course_partner_id)
VALUES('01GWK5JB12FJC16STAMWWBCOURSE', 'course DWH test', NULL, NULL, NULL, 0, '2023-07-19 17:49:25.803', '2023-07-19 17:49:25.803', -2147483642, NULL, NULL, NULL, NULL, NULL, NULL, NULL, 'COURSE_STATUS_NONE', '-2147483642', NULL, '01GWK5JB12FJC16STACOURSETYPE', NULL, false, NULL);

INSERT INTO public.course_access_paths (course_id,location_id,resource_path,created_at,updated_at) VALUES
    ('01GWK5JB12FJC16STAMWWBCOURSE','01GV032YZ8FA4JGEAR4XXQX6L4','-2147483642',now(),now()) ON CONFLICT DO NOTHING;


INSERT INTO public."class"
(class_id, "name", course_id, school_id, location_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GV032YZ8FA4JGEAR4XXQCLASS', 'class DWH test', '01GWK5JB12FJC16STAMWWBCOURSE', 'school_id', '01GV032YZ8FA4JGEAR4XXQX6L4', '2023-07-19 17:46:09.612', '2023-07-19 17:46:09.612', NULL, '-2147483642');

INSERT INTO public.class_member
(class_member_id, class_id, user_id, created_at, updated_at, deleted_at, resource_path, start_date, end_date)
VALUES('01GV032YZ8FA4JGEAR4XMEM1', '01GV032YZ8FA4JGEAR4XXQCLASS', '01GSX7KMWWED9ZZ79GDZ7ZX525', '2023-07-19 17:47:35.220', '2023-07-19 17:47:35.220', NULL, '-2147483642', '2023-07-19 17:47:35.220', '2023-07-19 17:47:35.220');
INSERT INTO public.class_member
(class_member_id, class_id, user_id, created_at, updated_at, deleted_at, resource_path, start_date, end_date)
VALUES('01GV032YZ8FA4JGEAR4XMEM2', '01GV032YZ8FA4JGEAR4XXQCLASS', '01GTZYX224982Z1X4MHZQW6DQ4', '2023-07-19 17:47:35.220', '2023-07-19 17:47:35.220', NULL, '-2147483642', '2023-07-19 17:47:35.220', '2023-07-19 17:47:35.220');

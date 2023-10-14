\connect mastermgmt;

INSERT INTO public.grade
(grade_id, name, is_archived, partner_internal_id, updated_at, created_at, deleted_at, resource_path, "sequence", remarks)
VALUES('01GV032YZ8FA4JGEAR4XXQX6L1', 'Grade DWH test', false, '1', '2023-07-18 11:51:38.657', '2023-07-18 11:51:38.657', NULL, '-2147483642', 1, '');


INSERT INTO public.academic_year
(academic_year_id, "name", start_date, end_date, created_at, updated_at, deleted_at, resource_path)
VALUES('01GV032YZ8FA4JGEAR4XXQX6L3', '2023 DWH Test',  now(),  now(), timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483642');


INSERT INTO public.course_academic_year
(course_id, academic_year_id, created_at, updated_at, deleted_at, resource_path)
VALUES('01GV032YZ8FA4JGEAR4XXQX6L3', '01GV032YZ8FA4JGEAR4XXQX6L3', timezone('utc'::text, now()), timezone('utc'::text, now()), NULL, '-2147483642');


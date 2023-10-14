INSERT INTO public.configs
(config_key, config_group, country, config_value, updated_at, created_at, deleted_at, resource_path)
VALUES
('specificCourseIDsForLesson', 'lesson', 'COUNTRY_MASTER', 'JPREP_COURSE_000000162,JPREP_COURSE_000000218,JPREP_COURSE_000000163', NOW(), NOW(), NULL, '-2147483647')
ON CONFLICT DO NOTHING;

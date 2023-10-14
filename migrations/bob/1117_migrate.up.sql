INSERT INTO public.configs
(config_key, config_group, country, config_value, updated_at, created_at, deleted_at)
VALUES
('domain_stag_bo_2147483647', 'lesson', 'COUNTRY_MASTER', 'https://staging-jprep-school-portal.web.app/', NOW(), NOW(), NULL),
('domain_stag_teacher_2147483647', 'lesson', 'COUNTRY_MASTER', 'https://staging-jprep-teacher.web.app/', NOW(), NOW(), NULL),
('domain_stag_learner_2147483647', 'lesson', 'COUNTRY_MASTER', 'https://staging-jprep-learner.web.app/', NOW(), NOW(), NULL)
ON CONFLICT DO NOTHING;
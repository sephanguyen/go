-- add constrain check config type
ALTER TABLE public.configuration_key DROP constraint if exists configuration_key_type_check;

ALTER TABLE public.configuration_key ADD CONSTRAINT configuration_key_type_check 
CHECK ((configuration_type = ANY (ARRAY['CONFIGURATION_TYPE_INTERNAL'::text, 'CONFIGURATION_TYPE_EXTERNAL'::text])));

-- init config_key default value
INSERT INTO public.configuration_key
(value_type, default_value, configuration_type, created_at, updated_at, config_key)
VALUES
('string', 'off', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'lesson.live_lesson.enable_live_lesson'),
('string', 'off', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'lesson.live_lesson.cloud_record'),
('string', 'off', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'lesson.lessonmgmt.zoom_selection'),
('boolean', 'true', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'lesson.lessonmgmt.allow_write_lesson'),        -- only Jprep don't need to write, they sync by api
('string', 'on', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'user.student_course.allow_input_student_course'), -- partner don't user ERP so need to maunal input
('string', 'off', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'user.enrollment.update_status_manual'),
('string', 'off', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'hcm.timesheet_management'),
('string', 'off', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'syllabus.learning_material.content_lo'),
('string', 'off', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'payment.order.enable_order_manager'),
('string', 'off', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'lesson.assigned_student_list'),
('string', 'off', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'lesson.lesson_report.enable_lesson_report'),
('string', 'off', 'CONFIGURATION_TYPE_EXTERNAL', NOW(), NOW(), 'user.authentication.ip_address_restriction'),
('string', '',  'CONFIGURATION_TYPE_EXTERNAL', NOW(), NOW(), 'user.authentication.allowed_ip_address'),
('string', 'off', 'CONFIGURATION_TYPE_EXTERNAL', NOW(), NOW(), 'syllabus.approve_grading'),
('string', '#566d8f', 'CONFIGURATION_TYPE_EXTERNAL', NOW(), NOW(), 'general.app_theme'),
('string', '', 'CONFIGURATION_TYPE_EXTERNAL', NOW(), NOW(), 'user.authentication.allowed_address_list'),
('string', 'https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQbWzBnqkpzJri-KyLdivc6ps9U3NP8XFTpOK-F4sKzAA&s', 'CONFIGURATION_TYPE_EXTERNAL', NOW(), NOW(), 'general.logo'),
('string', 'off', 'CONFIGURATION_TYPE_EXTERNAL', NOW(), NOW(), 'lesson.lessonmgmt.lesson_selection'),
('json', '', 'CONFIGURATION_TYPE_EXTERNAL', NOW(), NOW(), 'lesson.zoom.config'),
('boolean', 'false', 'CONFIGURATION_TYPE_EXTERNAL', NOW(), NOW(), 'lesson.zoom.is_enabled')
ON CONFLICT DO NOTHING;

-- add foreign key
ALTER TABLE public.internal_configuration_value DROP constraint if exists internal_configuration_value_key;

ALTER TABLE public.internal_configuration_value  
ADD constraint internal_configuration_value_key 
FOREIGN KEY (config_key) 
REFERENCES public.configuration_key(config_key);

ALTER TABLE public.external_configuration_value DROP constraint if exists external_configuration_value_key;

ALTER TABLE public.external_configuration_value  
ADD CONSTRAINT external_configuration_value_key
FOREIGN KEY (config_key) 
REFERENCES public.configuration_key(config_key);
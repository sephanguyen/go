INSERT INTO public.configuration_key
(value_type, default_value, configuration_type, created_at, updated_at, config_key)
VALUES
('string', 'on', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'user.student.enable_student_management'),
('string', 'on', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'user.staff.enable_staff_management'),
('string', 'on', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'user.user_group.enable_user_group_management')
ON CONFLICT DO NOTHING;

-- Disable in Jprep
UPDATE internal_configuration_value 
SET config_value = 'off'
WHERE config_key = 'user.student.enable_student_management' and resource_path ='-2147483647';

UPDATE internal_configuration_value 
SET config_value = 'off'
WHERE config_key = 'user.staff.enable_staff_management' and resource_path ='-2147483647';

UPDATE internal_configuration_value 
SET config_value = 'off'
WHERE config_key = 'user.user_group.enable_user_group_management' and resource_path ='-2147483647';
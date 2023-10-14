-- init config_key default value
INSERT INTO public.configuration_key
(value_type, default_value, configuration_type, created_at, updated_at, config_key)
VALUES
('string', 'off', 'CONFIGURATION_TYPE_INTERNAL', NOW(), NOW(), 'user.student_management.deactivate_parent')
ON CONFLICT DO NOTHING;
-- Disable Withus
UPDATE internal_configuration_value
SET config_value = 'false'
WHERE config_key = 'lesson.class_do.is_enabled' and resource_path ='-2147483624';

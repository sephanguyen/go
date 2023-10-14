-- Enable in kec-demo
UPDATE internal_configuration_value
SET config_value = 'true'
WHERE config_key = 'lesson.advanced_class_registration.is_enabled' and resource_path ='-2147483635';

------------------------------------------------------------------------------------------------------------------------

-- Enable in kec-demo
UPDATE internal_configuration_value
SET config_value = 'true'
WHERE config_key = 'lesson.attendance_status_filter.is_enabled' and resource_path ='-2147483635';

------------------------------------------------------------------------------------------------------------------------

-- Enable in kec-demo
UPDATE internal_configuration_value
SET config_value = 'true'
WHERE config_key = 'lesson.auto_filter_by_lesson_course_for_group_lessons.is_enabled' and resource_path ='-2147483635';

INSERT INTO public.configuration_key(
    config_key,
    value_type,
    default_value,
    configuration_type,
    created_at,
    updated_at
)
VALUES(
          'lesson.advanced_class_registration.is_enabled',
          'boolean',
          'false',
          'CONFIGURATION_TYPE_INTERNAL',
          NOW(),
          NOW()
      );

-- Enable in E2E
UPDATE internal_configuration_value
SET config_value = 'true'
WHERE config_key = 'lesson.advanced_class_registration.is_enabled' and resource_path ='-2147483639';

UPDATE internal_configuration_value
SET config_value = 'true'
WHERE config_key = 'lesson.advanced_class_registration.is_enabled' and resource_path ='-2147483638';

------------------------------------------------------------------------------------------------------------------------

INSERT INTO public.configuration_key(
    config_key,
    value_type,
    default_value,
    configuration_type,
    created_at,
    updated_at
)
VALUES(
          'lesson.attendance_status_filter.is_enabled',
          'boolean',
          'false',
          'CONFIGURATION_TYPE_INTERNAL',
          NOW(),
          NOW()
      );

-- Enable in E2E
UPDATE internal_configuration_value
SET config_value = 'true'
WHERE config_key = 'lesson.attendance_status_filter.is_enabled' and resource_path ='-2147483639';

UPDATE internal_configuration_value
SET config_value = 'true'
WHERE config_key = 'lesson.attendance_status_filter.is_enabled' and resource_path ='-2147483638';

------------------------------------------------------------------------------------------------------------------------

INSERT INTO public.configuration_key(
    config_key,
    value_type,
    default_value,
    configuration_type,
    created_at,
    updated_at
)
VALUES(
          'lesson.auto_filter_by_lesson_course_for_group_lessons.is_enabled',
          'boolean',
          'false',
          'CONFIGURATION_TYPE_INTERNAL',
          NOW(),
          NOW()
      );

-- Enable in E2E
UPDATE internal_configuration_value
SET config_value = 'true'
WHERE config_key = 'lesson.auto_filter_by_lesson_course_for_group_lessons.is_enabled' and resource_path ='-2147483639';

UPDATE internal_configuration_value
SET config_value = 'true'
WHERE config_key = 'lesson.auto_filter_by_lesson_course_for_group_lessons.is_enabled' and resource_path ='-2147483638';

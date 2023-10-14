INSERT INTO public.configuration_key(
    config_key,
    value_type,
    default_value,
    configuration_type,
    created_at,
    updated_at
)
VALUES(
          'lesson.class_do.is_enabled',
          'boolean',
          'false',
          'CONFIGURATION_TYPE_INTERNAL',
          NOW(),
          NOW()
      );

-- Enable in Manabie
UPDATE internal_configuration_value
SET config_value = 'true'
WHERE config_key = 'lesson.class_do.is_enabled' and resource_path ='-2147483648';

-- Enable in E2E
UPDATE internal_configuration_value
SET config_value = 'true'
WHERE config_key = 'lesson.class_do.is_enabled' and resource_path ='-2147483639';

UPDATE internal_configuration_value
SET config_value = 'true'
WHERE config_key = 'lesson.class_do.is_enabled' and resource_path ='-2147483638';

-- Enable in Withus
UPDATE internal_configuration_value
SET config_value = 'true'
WHERE config_key = 'lesson.class_do.is_enabled' and resource_path ='-2147483624';

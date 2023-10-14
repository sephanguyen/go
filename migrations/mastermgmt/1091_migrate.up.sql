INSERT INTO public.configuration_key(
    config_key,
    value_type,
    default_value,
    configuration_type,
    created_at,
    updated_at
)
VALUES(
    'lesson.lesson_allocation.enable_lesson_allocation',
    'string',
    'off',
    'CONFIGURATION_TYPE_INTERNAL',
    NOW(),
    NOW()
);

-- Enable in Manabie
UPDATE internal_configuration_value
SET config_value = 'on'
WHERE config_key = 'lesson.lesson_allocation.enable_lesson_allocation' and resource_path ='-2147483648';

-- Enable in E2E-Tokyo
UPDATE internal_configuration_value
SET config_value = 'on'
WHERE config_key = 'lesson.lesson_allocation.enable_lesson_allocation' and resource_path ='-2147483639';

-- Enable in PROD KEC ERP demo
UPDATE internal_configuration_value
SET config_value = 'on'
WHERE config_key = 'lesson.lesson_allocation.enable_lesson_allocation' and resource_path ='-2147483635';

-- Enable in PROD KEC ERP
UPDATE internal_configuration_value
SET config_value = 'on'
WHERE config_key = 'lesson.lesson_allocation.enable_lesson_allocation' and resource_path ='-2147483642';

-- Enable in KEC-UAT phase 2
UPDATE internal_configuration_value
SET config_value = 'on'
WHERE config_key = 'lesson.lesson_allocation.enable_lesson_allocation' and resource_path ='-2147483623';

-- Enable in Manabie Demo (For JP Sale team) - ERP
UPDATE internal_configuration_value
SET config_value = 'on'
WHERE config_key = 'lesson.lesson_allocation.enable_lesson_allocation' and resource_path ='-2147483628';

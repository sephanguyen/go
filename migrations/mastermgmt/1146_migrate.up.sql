INSERT INTO public.configuration_key(
    config_key,
    value_type,
    default_value,
    configuration_type,
    created_at,
    updated_at
)
VALUES(
    'user.student_management.erp_modules',
    'string',
    'off',
    'CONFIGURATION_TYPE_INTERNAL',
    NOW(),
    NOW()
);

UPDATE internal_configuration_value
SET config_value = 'on'
WHERE 
    config_key = 'user.student_management.erp_modules' AND 
    resource_path IN ('-2147483635', '-2147483642', '-2147483628', '-2147483623');

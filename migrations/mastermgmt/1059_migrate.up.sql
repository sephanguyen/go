INSERT INTO public.configuration_key(
    config_key,
    value_type,
    default_value,
    configuration_type,
    created_at,
    updated_at
)
VALUES(
    'syllabus.to_review.assignment_type',
    'boolean',
    'true',
    'CONFIGURATION_TYPE_INTERNAL',
    NOW(),
    NOW()
);

-- Disabled for managara-base and managara-hs
UPDATE internal_configuration_value
SET config_value = 'false'
WHERE config_key = 'syllabus.to_review.assignment_type' and resource_path IN ('-2147483630', '-2147483629');

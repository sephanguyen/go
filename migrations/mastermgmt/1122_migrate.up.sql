INSERT INTO public.configuration_key(
    config_key,
    value_type,
    default_value,
    configuration_type,
    created_at,
    updated_at
)
VALUES(
    'user.dashboard.enrolled_students_widget',
    'string',
    'off',
    'CONFIGURATION_TYPE_INTERNAL',
    NOW(),
    NOW()
);




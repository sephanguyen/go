INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'user.auth.allow_change_password_on_learner', 
    'string', 
    'on', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);

-- Disable in Withus Managara-HS
UPDATE internal_configuration_value 
SET config_value = 'off'
WHERE config_key = 'user.auth.allow_change_password_on_learner' and resource_path ='-2147483629';

-- Disable in Withus Managara-Base
UPDATE internal_configuration_value 
SET config_value = 'off'
WHERE config_key = 'user.auth.allow_change_password_on_learner' and resource_path ='-2147483630';

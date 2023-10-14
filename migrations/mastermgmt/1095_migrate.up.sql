INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'communication.notification.enable_notification', 
    'string', 
    'on', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);

-- Disable in kec-demo
UPDATE internal_configuration_value 
SET config_value = 'off'
WHERE config_key = 'communication.notification.enable_notification' and resource_path ='-2147483635';

-- Disable in kec
UPDATE internal_configuration_value 
SET config_value = 'off'
WHERE config_key = 'communication.notification.enable_notification' and resource_path ='-2147483642';

-- Disable in kec-test (KEC-UAT P2) 
UPDATE internal_configuration_value 
SET config_value = 'off'
WHERE config_key = 'communication.notification.enable_notification' and resource_path ='-2147483623';


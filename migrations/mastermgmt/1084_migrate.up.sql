INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'communication.dashboard.enable_dashboard', 
    'string', 
    'off', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);

-- Enable in E2E-Tokyo
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'communication.dashboard.enable_dashboard' and resource_path ='-2147483639';

-- Enable in Manabie
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'communication.dashboard.enable_dashboard' and resource_path ='-2147483648';

-- Enable in KEC
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'communication.dashboard.enable_dashboard' and resource_path ='-2147483642';

-- Enable in GA
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'communication.dashboard.enable_dashboard' and resource_path ='-2147483644';

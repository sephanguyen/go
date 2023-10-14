--Default value for brightcove config use tokyo config --
INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES
(
    'mastermgmt.brightcoveconfig.account_id', 
    'string', 
    '6064018595001',
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
),
(
    'mastermgmt.brightcoveconfig.client_id', 
    'string', 
    '7f7d1f2e-9a66-4cf5-8187-95aabd9ccaa8', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
),
(
    'mastermgmt.brightcoveconfig.profile', 
    'string', 
    'Asia-PREMIUM (96-1500)', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);

-- JPREP --
UPDATE public.internal_configuration_value
SET config_value='6223705322001' 
WHERE config_key='mastermgmt.brightcoveconfig.account_id' AND resource_path='-2147483647';

UPDATE public.internal_configuration_value
SET config_value='d3b712c7-b491-4ac1-81ee-ce564c3c42f0' 
WHERE config_key='mastermgmt.brightcoveconfig.client_id' AND resource_path='-2147483647';

UPDATE public.internal_configuration_value
SET config_value='multi-platform-standard-static' 
WHERE config_key='mastermgmt.brightcoveconfig.profile' AND resource_path='-2147483647';

-- Synersia --
UPDATE public.internal_configuration_value
SET config_value='6228002151001' 
WHERE config_key='mastermgmt.brightcoveconfig.account_id' AND resource_path='-2147483646';

UPDATE public.internal_configuration_value
SET config_value='b3319002-aa19-4cfd-a2e0-6898987a4539' 
WHERE config_key='mastermgmt.brightcoveconfig.client_id' AND resource_path='-2147483646';

UPDATE public.internal_configuration_value
SET config_value='multi-platform-standard-static' 
WHERE config_key='mastermgmt.brightcoveconfig.profile' AND resource_path='-2147483646';

-- Renseikai --
UPDATE public.internal_configuration_value
SET config_value='6248889468001' 
WHERE config_key='mastermgmt.brightcoveconfig.account_id' AND resource_path='-2147483645';

UPDATE public.internal_configuration_value
SET config_value='7c38ce89-153a-43d8-8a21-e67616a10b1e' 
WHERE config_key='mastermgmt.brightcoveconfig.client_id' AND resource_path='-2147483645';

UPDATE public.internal_configuration_value
SET config_value='multi-platform-standard-dynamic' 
WHERE config_key='mastermgmt.brightcoveconfig.profile' AND resource_path='-2147483645';

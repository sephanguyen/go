INSERT INTO public.configuration_key(
    config_key,
    value_type, 
    default_value, 
    configuration_type, 
    created_at, 
    updated_at
)
VALUES(
    'entryexit.entryexitmgmt.enable_entryexit_manager', 
    'string', 
    'off', 
    'CONFIGURATION_TYPE_INTERNAL', 
    NOW(), 
    NOW()
);

-- Enable in Manabie
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483648';

-- Enable in Synersia
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483646';

-- Enable in Renseikai
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483645';

-- Enable in E2E
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483644';

-- Enable in BestCo
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483643';

-- Enable in AIC
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483641';

-- Enable in NSG
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483640';

-- Enable in E2E-Tokyo
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483639';

-- Enable in E2E-HCM
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483638';

-- Enable in Manabie-demo
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483637';

-- Enable in Manabie Internal
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483634';

-- Enable in Manabie Tech
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483633';

-- Enable in Manabie Kael
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483632';

-- Enable in Eishinkan
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483631';

-- Enable in Managara Base
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483630';

-- Enable in Managara High School
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483629';

-- Enable in E2E Architecture
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='100000';
-- Enable in PROD KEC ERP demo
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'communication.dashboard.enable_dashboard' and resource_path ='-2147483635';

-- Enable in PROD Bestco
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'communication.dashboard.enable_dashboard' and resource_path ='-2147483643';

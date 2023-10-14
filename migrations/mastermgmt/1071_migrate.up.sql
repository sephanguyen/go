-- Enable in withus-juku
UPDATE internal_configuration_value 
SET config_value = 'on'
WHERE config_key = 'entryexit.entryexitmgmt.enable_entryexit_manager' and resource_path ='-2147483624';

UPDATE internal_configuration_value 
SET config_value = 'on' 
WHERE resource_path = ANY ('{-2147483630, -2147483629}') 
AND config_key  ='user.enrollment.update_status_manual';
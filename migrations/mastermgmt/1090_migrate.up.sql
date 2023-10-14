-- Disable in Withus Juku
UPDATE internal_configuration_value 
SET config_value = 'off'
WHERE config_key = 'user.auth.allow_change_password_on_learner' and resource_path ='-2147483624';

-- Disable in Renseikai
UPDATE internal_configuration_value 
SET config_value = 'off'
WHERE config_key = 'user.auth.allow_change_password_on_learner' and resource_path ='-2147483645';
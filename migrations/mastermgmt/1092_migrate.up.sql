UPDATE external_configuration_value
SET config_value = '{"ipv4": [], "ipv6": []}'
WHERE
        config_key = 'user.authentication.allowed_ip_address' AND
        config_value = '[]';

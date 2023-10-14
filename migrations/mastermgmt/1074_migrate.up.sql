---------- MIGRATE SET DEFAULT VALUE ----------
UPDATE configuration_key
SET default_value = '{"ipv4": [], "ipv6": []}'
WHERE
    config_key = 'user.authentication.allowed_ip_address';

UPDATE external_configuration_value
SET config_value = '{"ipv4": [], "ipv6": []}'
WHERE
        config_key = 'user.authentication.allowed_ip_address'
    AND (config_value = '' OR config_value is NULL);


---------- MIGRATE ADD FUNCTION ADD IP ADDRESS ----------
DROP FUNCTION IF EXISTS ip_restriction_add_ip_address;
CREATE FUNCTION ip_restriction_add_ip_address(resource_path_param TEXT, ip_address_type TEXT, ip_address TEXT) RETURNS VOID AS $$
DECLARE
  value JSONB;
BEGIN
  SELECT config_value::jsonb INTO value
  FROM external_configuration_value
  WHERE resource_path = resource_path_param
    AND config_key = 'user.authentication.allowed_ip_address';
  
  IF ip_address_type = 'ipv4' THEN -- append for ipv4
    -- append new value
    value := jsonb_set(value, '{ipv4}', value->'ipv4' || jsonb_build_array(ip_address::text));
    -- remove duplicated values
    value := jsonb_set(value, '{ipv4}', (SELECT jsonb_agg(DISTINCT elem) FROM jsonb_array_elements(value->'ipv4') AS elem));

  ELSE -- append for ipv6
    -- append new value
    value := jsonb_set(value, '{ipv6}', value->'ipv6' || jsonb_build_array(ip_address::text));
    -- remove duplicated values
    value := jsonb_set(value, '{ipv6}', (SELECT jsonb_agg(DISTINCT elem) FROM jsonb_array_elements(value->'ipv6') AS elem));
  END IF;
  
  UPDATE external_configuration_value
  SET config_value = value::text
  WHERE resource_path = resource_path_param AND config_key = 'user.authentication.allowed_ip_address';
END;
$$ LANGUAGE plpgsql;



---------- MIGRATE ADD FUNCTION REMOVE IP ADDRESS ----------
DROP FUNCTION IF EXISTS ip_restriction_remove_ip_address;
CREATE OR REPLACE FUNCTION ip_restriction_remove_ip_address(resource_path_param TEXT, ip_address_type TEXT, ip_address TEXT) RETURNS VOID AS $$
DECLARE
  value JSONB;
BEGIN
  SELECT config_value::jsonb INTO value
  FROM external_configuration_value
  WHERE resource_path = resource_path_param
    AND config_key = 'user.authentication.allowed_ip_address';
  
  -- remove ip address by select existed array and ignore value want to be removed
  IF ip_address_type = 'ipv4' THEN
    value = jsonb_set(
      value,
      '{ipv4}',
      COALESCE((
          SELECT jsonb_agg(DISTINCT elem)
          FROM jsonb_array_elements(value->'ipv4') AS elem
          WHERE trim(elem::text, '"') <> ip_address)
        , '[]'::jsonb -- default value if the result of the select statement is null
      )
    );
  ELSE
    value = jsonb_set(
      value,
      '{ipv6}',
      COALESCE((
          SELECT jsonb_agg(DISTINCT elem)
          FROM jsonb_array_elements(value->'ipv6') AS elem
          WHERE trim(elem::text, '"') <> ip_address)
        , '[]'::jsonb
      )
    );
  END IF;

  UPDATE external_configuration_value
  SET config_value = value::text
  WHERE resource_path = resource_path_param
    AND config_key = 'user.authentication.allowed_ip_address';
END;
$$ LANGUAGE plpgsql;

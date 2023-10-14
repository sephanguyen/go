-- when insert a new config key, create value for all existing partner by default value
CREATE OR REPLACE RULE INIT_CONFIG_INTERNAL_VALUE_FOR_NEW_KEY
AS ON INSERT TO configuration_key
do also 
	INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) 
	select uuid_generate_v4(), new.config_key, new.value_type, now(), now(), new.default_value, resource_path
	from organizations o 
	where new.configuration_type = 'CONFIGURATION_TYPE_INTERNAL';
	
CREATE OR REPLACE RULE INIT_CONFIG_EXTERNAL_VALUE_FOR_NEW_KEY
AS ON INSERT TO configuration_key
do also 
	INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) 
	select uuid_generate_v4(), new.config_key, new.value_type, now(), now(), new.default_value, resource_path
	from organizations o 
	where new.configuration_type = 'CONFIGURATION_TYPE_EXTERNAL';

alter table configuration_key ENABLE ALWAYS RULE INIT_CONFIG_INTERNAL_VALUE_FOR_NEW_KEY;
alter table configuration_key ENABLE ALWAYS RULE INIT_CONFIG_EXTERNAL_VALUE_FOR_NEW_KEY;

-- whn insert a new partner, create value for all existing config_key by default value
CREATE OR REPLACE RULE INIT_CONFIG_INTERNAL_VALUE_FOR_NEW_PARTNER
AS ON INSERT TO organizations
do also 
	INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) 
	select uuid_generate_v4(), ck.config_key, ck.value_type, now(), now(), ck.default_value, new.resource_path
	from configuration_key ck 
	where ck.configuration_type = 'CONFIGURATION_TYPE_INTERNAL';
	
CREATE OR REPLACE RULE INIT_CONFIG_EXTERNAL_VALUE_FOR_NEW_PARTNER
AS ON INSERT TO organizations
do also 
	INSERT INTO public."external_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path) 
	select uuid_generate_v4(), ck.config_key, ck.value_type, now(), now(), ck.default_value, new.resource_path
	from configuration_key ck 
	where ck.configuration_type = 'CONFIGURATION_TYPE_EXTERNAL';

alter table organizations ENABLE ALWAYS RULE INIT_CONFIG_INTERNAL_VALUE_FOR_NEW_PARTNER;
alter table organizations ENABLE ALWAYS RULE INIT_CONFIG_EXTERNAL_VALUE_FOR_NEW_PARTNER;

-- don't need to check config_value insert to all partner anymore, 2 above rule cover that
drop rule if exists CHECK_CONFIG_ALL_PARTNER on internal_configuration_value;
drop rule if exists CHECK_CONFIG_ALL_PARTNER on external_configuration_value;
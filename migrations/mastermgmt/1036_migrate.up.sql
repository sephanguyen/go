-- add rule to force insert config value for all partner
CREATE OR REPLACE RULE CHECK_CONFIG_ALL_PARTNER 
AS ON INSERT TO internal_configuration_value
do also 
		delete from internal_configuration_value cv
		where cv.config_key = new.config_key 
			and config_key in (select icv.config_key 
								from internal_configuration_value icv 
									join organizations o on o.resource_path = icv.resource_path 
								group by icv.config_key 
								having count(*)< (select count(*) from organizations o2));

CREATE OR REPLACE RULE CHECK_CONFIG_ALL_PARTNER 
AS ON INSERT TO external_configuration_value 
do also 
		delete from external_configuration_value cv
		where cv.config_key = new.config_key 
			and config_key in (select ecv.config_key 
								from external_configuration_value ecv 
									join organizations o on o.resource_path = ecv.resource_path 
								group by ecv.config_key 
								having count(*)< (select count(*) from organizations o2));
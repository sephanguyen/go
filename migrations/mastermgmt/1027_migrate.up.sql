ALTER TABLE external_configuration
ADD CONSTRAINT config_key_resource_unique UNIQUE(config_key, resource_path);

CREATE TABLE IF NOT EXISTS public.location_configuration_value (
    location_config_id TEXT NOT NULL,
    config_key TEXT NOT NULL,
    location_id TEXT NOT NULL,
    config_value TEXT,
    config_value_type TEXT NOT NULL DEFAULT 'string'::text,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT location_configuration_value__pk PRIMARY KEY (location_config_id),
    CONSTRAINT location_configuration_value_fk FOREIGN KEY (config_key, resource_path) REFERENCES public.external_configuration_value (config_key, resource_path),
    CONSTRAINT location_configuration_value_resource_unique UNIQUE(config_key, location_id, resource_path)
);

CREATE POLICY rls_location_configuration_value ON "location_configuration_value"
USING (permission_check(resource_path, 'location_configuration_value')) WITH CHECK (permission_check(resource_path, 'location_configuration_value'));
CREATE POLICY rls_location_configuration_value_restrictive ON "location_configuration_value" AS RESTRICTIVE
USING (permission_check(resource_path, 'location_configuration_value'))WITH CHECK (permission_check(resource_path, 'location_configuration_value'));

ALTER TABLE "location_configuration_value" ENABLE ROW LEVEL security;
ALTER TABLE "location_configuration_value" FORCE ROW LEVEL security;

CREATE TABLE IF NOT EXISTS public.location_configuration_value_v2 (
    location_config_id TEXT NOT NULL,
    config_key TEXT NOT NULL,
    location_id TEXT NOT NULL,
    config_value TEXT,
    config_value_type TEXT NOT NULL DEFAULT 'string'::text,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT location_configuration_value__pk_v2 PRIMARY KEY (location_config_id),
    CONSTRAINT location_configuration_value_resource_unique_v2 UNIQUE(config_key, location_id, resource_path)
);

CREATE POLICY rls_location_configuration_value_v2 ON "location_configuration_value_v2" using (permission_check(resource_path, 'location_configuration_value_v2')) with check (permission_check(resource_path, 'location_configuration_value_v2'));
CREATE POLICY rls_location_configuration_value_v2_restrictive ON "location_configuration_value_v2" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'location_configuration_value_v2')) with check (permission_check(resource_path, 'location_configuration_value_v2'));

ALTER TABLE "location_configuration_value_v2" ENABLE ROW LEVEL security;
ALTER TABLE "location_configuration_value_v2" FORCE ROW LEVEL security;

ALTER TABLE public.configuration_key DROP constraint if exists configuration_key_type_check;

ALTER TABLE public.configuration_key ADD CONSTRAINT configuration_key_type_check 
CHECK ((configuration_type = ANY (ARRAY[
    'CONFIGURATION_TYPE_INTERNAL'::text, 
    'CONFIGURATION_TYPE_EXTERNAL'::text,
    'CONFIGURATION_TYPE_LOCATION_EXTERNAL'::text
])));

/* configuration_key */
CREATE TABLE IF NOT EXISTS public.configuration_key (
    config_key TEXT NOT NULL,
    value_type TEXT NOT NULL,
    default_value TEXT,
    configuration_type TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT config_key__pk PRIMARY KEY (config_key)
);

/* internal_configuration_value */

CREATE TABLE IF NOT EXISTS public.internal_configuration_value (
    configuration_id TEXT NOT NULL,
    config_key TEXT NOT NULL,
    config_value TEXT,
    config_value_type TEXT NOT NULL DEFAULT 'string'::text,
    last_editor text,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT internal_configuration_value__pk PRIMARY KEY (configuration_id),
    CONSTRAINT internal_configuration_value_resource_unique UNIQUE(config_key, resource_path)
);


CREATE POLICY rls_internal_configuration_value ON "internal_configuration_value"
USING (permission_check(resource_path, 'internal_configuration_value')) WITH CHECK (permission_check(resource_path, 'internal_configuration_value'));
CREATE POLICY rls_internal_configuration_value_restrictive ON "internal_configuration_value" AS RESTRICTIVE
USING (permission_check(resource_path, 'internal_configuration_value'))WITH CHECK (permission_check(resource_path, 'internal_configuration_value'));

ALTER TABLE "internal_configuration_value" ENABLE ROW LEVEL security;
ALTER TABLE "internal_configuration_value" FORCE ROW LEVEL security;

/* external_configuration_value */

CREATE TABLE IF NOT EXISTS public.external_configuration_value (
    configuration_id TEXT NOT NULL,
    config_key TEXT NOT NULL,
    config_value TEXT,
    config_value_type TEXT NOT NULL DEFAULT 'string'::text,
    last_editor text,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT external_configuration_value__pk PRIMARY KEY (configuration_id),
    CONSTRAINT external_configuration_value_resource_unique UNIQUE(config_key, resource_path)
);


CREATE POLICY rls_external_configuration_value ON "external_configuration_value"
USING (permission_check(resource_path, 'external_configuration_value')) WITH CHECK (permission_check(resource_path, 'external_configuration_value'));
CREATE POLICY rls_external_configuration_value_restrictive ON "external_configuration_value" AS RESTRICTIVE
USING (permission_check(resource_path, 'external_configuration_value'))WITH CHECK (permission_check(resource_path, 'external_configuration_value'));

ALTER TABLE "external_configuration_value" ENABLE ROW LEVEL security;
ALTER TABLE "external_configuration_value" FORCE ROW LEVEL security;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";


INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
values
(uuid_generate_v4(), 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483648') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
values
(uuid_generate_v4(), 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483647') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
values
(uuid_generate_v4(), 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483646') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
values
(uuid_generate_v4(), 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483645') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
values
(uuid_generate_v4(), 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483644') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
values
(uuid_generate_v4(), 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483643') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
values
(uuid_generate_v4(), 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483642') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
values
(uuid_generate_v4(), 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483641') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
values
(uuid_generate_v4(), 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483640') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
values
(uuid_generate_v4(), 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483636') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
values
(uuid_generate_v4(), 'hcm.timesheet_management', 'string', now(), now(), 'on', '-2147483635') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
values
(uuid_generate_v4(), 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483631') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
values
(uuid_generate_v4(), 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483630') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;
INSERT INTO public."internal_configuration_value" (configuration_id, config_key, config_value_type, created_at, updated_at, config_value, resource_path)
values
(uuid_generate_v4(), 'hcm.timesheet_management', 'string', now(), now(), 'off', '-2147483629') ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO NOTHING;
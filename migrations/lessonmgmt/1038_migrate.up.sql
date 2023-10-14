--

CREATE TABLE public.configs (
    config_key text NOT NULL,
    config_group text NOT NULL,
    country text NOT NULL,
    config_value text NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamptz NULL,
	resource_path text NULL DEFAULT autofillresourcepath(),
    CONSTRAINT config_pk PRIMARY KEY (config_key, config_group, country)

);

CREATE POLICY rls_configs ON "configs" USING (permission_check(resource_path, 'configs')) WITH CHECK (permission_check(resource_path, 'configs'));
CREATE POLICY rls_configs_restrictive ON "configs" AS RESTRICTIVE FOR ALL TO PUBLIC USING (permission_check(resource_path, 'configs')) WITH CHECK (permission_check(resource_path, 'configs'));

ALTER TABLE "configs" ENABLE ROW LEVEL security;
ALTER TABLE "configs" FORCE ROW LEVEL security;
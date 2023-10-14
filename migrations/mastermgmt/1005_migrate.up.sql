CREATE TABLE IF NOT EXISTS public.external_configuration(
  config_id text not null,
  config_key text not null,
  config_value text null,
	created_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
	updated_at timestamp with time zone NOT NULL DEFAULT (now() at time zone 'utc'),
	deleted_at timestamp with time zone,
  resource_path text null DEFAULT autofillresourcepath(),
  CONSTRAINT external_config_pk PRIMARY KEY(config_id)
);

CREATE POLICY rls_external_configuration ON "external_configuration" using (permission_check(resource_path, 'external_configuration')) with check (permission_check(resource_path, 'external_configuration'));

ALTER TABLE "external_configuration" ENABLE ROW LEVEL security;
ALTER TABLE "external_configuration" FORCE ROW LEVEL security;
CREATE TABLE IF NOT EXISTS public.configuration(
  config_id text not null,
  config_key text not null,
  config_value text null,
	created_at timestamp with time zone NOT NULL,
	updated_at timestamp with time zone NOT NULL,
	deleted_at timestamp with time zone,
  resource_path text null DEFAULT autofillresourcepath(),
  CONSTRAINT config_pk PRIMARY KEY(config_id)
);

CREATE POLICY rls_configuration ON "configuration" using (permission_check(resource_path, 'configuration')) with check (permission_check(resource_path, 'configuration'));

ALTER TABLE "configuration" ENABLE ROW LEVEL security;
ALTER TABLE "configuration" FORCE ROW LEVEL security;
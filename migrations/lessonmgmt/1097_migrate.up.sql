CREATE TABLE public.organizations (
	organization_id text NOT NULL,
	tenant_id text NULL,
	"name" text NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	domain_name text NULL,
	logo_url text NULL,
	country text NULL,
	created_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	updated_at timestamptz NOT NULL DEFAULT timezone('utc'::text, now()),
	deleted_at timestamptz NULL,
	description text NULL,
	CONSTRAINT organization__domain_name__un UNIQUE (domain_name),
	CONSTRAINT organizations__pk PRIMARY KEY (organization_id),
	CONSTRAINT organizations__tenant_id__un UNIQUE (tenant_id)
);
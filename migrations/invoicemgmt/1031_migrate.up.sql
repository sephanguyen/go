CREATE TABLE IF NOT EXISTS public.organizations
(
    organization_id text NOT NULL,
    tenant_id text,
    name text,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),
    country text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT organizations__pk PRIMARY KEY (organization_id),
    CONSTRAINT organizations__tenant_id__un UNIQUE (tenant_id)
)

CREATE TABLE IF NOT EXISTS public.organizations
(
    organization_id text ,
    tenant_id text,
    name text,
    resource_path text,
    domain_name text,
    logo_url text,
    country text,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    CONSTRAINT organizations__pk PRIMARY KEY (organization_id)
);

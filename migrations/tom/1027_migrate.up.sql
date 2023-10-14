CREATE TABLE IF NOT EXISTS public.locations
(
    location_id text NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at timestamp with time zone NOT NULL DEFAULT timezone('utc'::text, now()),
    deleted_at timestamp with time zone,
    resource_path text,
    location_type text,
    partner_internal_id text,
    partner_internal_parent_id text,
    parent_location_id text,
    is_archived boolean NOT NULL,
    access_path text,
    CONSTRAINT locations_pkey PRIMARY KEY (location_id)
);

CREATE TABLE IF NOT EXISTS public.location_types
(
    location_type_id text NOT NULL,
    name text NOT NULL,
    display_name text,
    parent_name text,
    parent_location_type_id text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text,
    is_archived boolean NOT NULL,
    CONSTRAINT location_types_pkey PRIMARY KEY (location_type_id)
);

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
)
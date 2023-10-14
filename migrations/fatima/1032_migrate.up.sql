CREATE TABLE public.location (
    id text NOT NULL,
    name text,
    location_type text,
    parent_location_id text,
    partner_internal_id text,
    partner_internal_parent_id text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT location_pk PRIMARY KEY (id)
);

CREATE POLICY rls_location ON "location" USING (permission_check(resource_path, 'location')) WITH CHECK (permission_check(resource_path, 'location'));

ALTER TABLE "location" ENABLE ROW LEVEL security;
ALTER TABLE "location" FORCE ROW LEVEL security;

CREATE TABLE public.locations (
    location_id text NOT NULL,
    name text,
    location_type text,
    parent_location_id text,
    partner_internal_id text,
    partner_internal_parent_id text,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone,
    is_archived boolean NOT NULL DEFAULT false,
    access_path text,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT location_pk PRIMARY KEY (location_id)
);

CREATE POLICY rls_locations ON "locations" USING (permission_check(resource_path, 'locations')) WITH CHECK (permission_check(resource_path, 'locations'));

ALTER TABLE "locations" ENABLE ROW LEVEL security;
ALTER TABLE "locations" FORCE ROW LEVEL security;

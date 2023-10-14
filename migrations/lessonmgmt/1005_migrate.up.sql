-- CREATE TABLE public.location_types;
CREATE TABLE IF NOT EXISTS public.location_types (
	location_type_id text NOT NULL,
	"name" text NOT NULL,
	display_name text NULL,
	parent_name text NULL,
	parent_location_type_id text NULL,
	updated_at timestamptz NOT NULL,
	created_at timestamptz NOT NULL,
	deleted_at timestamptz NULL,
	resource_path text NOT NULL DEFAULT autofillresourcepath(),
	is_archived bool NOT NULL DEFAULT false,
	CONSTRAINT location_types_pkey PRIMARY KEY (location_type_id),
	CONSTRAINT unique__location_type_name_resource_path UNIQUE (name, resource_path),
	CONSTRAINT location_type_id_fk FOREIGN KEY (parent_location_type_id) REFERENCES public.location_types(location_type_id)
);

CREATE policy rls_location_types ON "location_types" USING (permission_check(resource_path, 'location_types')) WITH CHECK (permission_check(resource_path, 'location_types'));
CREATE POLICY rls_location_types_restrictive ON "location_types" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'location_types')) with check (permission_check(resource_path, 'location_types'));


ALTER TABLE "location_types" ENABLE ROW LEVEL SECURITY;
ALTER TABLE "location_types" FORCE ROW LEVEL SECURITY;

-- CREATE TABLE public.locations;
CREATE TABLE IF NOT EXISTS public.locations (
    location_id text NOT NULL,
    name text,
    location_type text,
    parent_location_id text,
    partner_internal_id text,
    partner_internal_parent_id text,
    updated_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    deleted_at timestamptz,
    is_archived boolean NOT NULL DEFAULT false,
    access_path text,
    resource_path text DEFAULT autofillresourcepath(),
    CONSTRAINT location_pk PRIMARY KEY (location_id)
);

CREATE POLICY rls_locations ON "locations" USING (permission_check(resource_path, 'locations')) WITH CHECK (permission_check(resource_path, 'locations'));
CREATE POLICY rls_locations_restrictive ON "locations" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'locations')) with check (permission_check(resource_path, 'locations'));

ALTER TABLE "locations" ENABLE ROW LEVEL security;
ALTER TABLE "locations" FORCE ROW LEVEL security;


ALTER TABLE ONLY public.locations ADD COLUMN IF NOT EXISTS location_type text NULL;
ALTER TABLE ONLY public.locations ADD COLUMN IF NOT EXISTS is_archived boolean NOT NULL DEFAULT false;
ALTER TABLE ONLY public.locations ADD COLUMN IF NOT EXISTS parent_location_id text NULL;

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
	"level" int4 NULL DEFAULT 0,

	CONSTRAINT location_types_pkey PRIMARY KEY (location_type_id),
	CONSTRAINT unique__location_type_name_resource_path UNIQUE (name, resource_path)
);

CREATE POLICY rls_location_types ON "location_types" using (permission_check(resource_path, 'location_types')) with check (permission_check(resource_path, 'location_types'));
CREATE POLICY rls_location_types_restrictive ON "location_types" AS RESTRICTIVE TO PUBLIC using (permission_check(resource_path, 'location_types')) with check (permission_check(resource_path, 'location_types'));

ALTER TABLE "location_types" ENABLE ROW LEVEL security;
ALTER TABLE "location_types" FORCE ROW LEVEL security;

ALTER TABLE "location" DISABLE ROW LEVEL security;
DROP POLICY IF EXISTS rls_location ON location ;
ALTER TABLE public.product_location DROP CONSTRAINT fk_location_id;

ALTER TABLE public.location RENAME TO locations;
ALTER TABLE public.locations
    RENAME id TO location_id;
ALTER TABLE public.product_location ADD CONSTRAINT fk_location_id FOREIGN KEY(location_id) REFERENCES locations(location_id);
ALTER TABLE locations ADD COLUMN is_archived boolean NOT NULL DEFAULT false;
ALTER TABLE locations ADD COLUMN access_path text;

CREATE POLICY rls_locations ON "locations" using (permission_check(resource_path, 'locations')) with check (permission_check(resource_path, 'locations'));

ALTER TABLE "locations" ENABLE ROW LEVEL security;
ALTER TABLE "locations" FORCE ROW LEVEL security;

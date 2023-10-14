ALTER TABLE locations ADD COLUMN IF NOT EXISTS location_type  TEXT DEFAULT 'LOCATION_TYPE_CENTER';

ALTER TABLE locations ADD COLUMN IF NOT EXISTS partner_internal_id TEXT;

ALTER TABLE locations ADD COLUMN IF NOT EXISTS partner_internal_parent_id TEXT;

ALTER TABLE locations ADD COLUMN IF NOT EXISTS parent_id TEXT;

ALTER TABLE IF EXISTS locations ADD CONSTRAINT location_id_fk FOREIGN KEY (parent_id) REFERENCES locations(location_id);

CREATE POLICY rls_locations ON "locations" using (permission_check(resource_path, 'locations')) with check (permission_check(resource_path, 'locations'));	

ALTER TABLE "locations" ENABLE ROW LEVEL security;	
ALTER TABLE "locations" FORCE ROW LEVEL security;
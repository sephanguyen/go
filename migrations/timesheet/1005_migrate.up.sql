CREATE TABLE IF NOT EXISTS public.locations
(
    location_id                TEXT                                                          NOT NULL PRIMARY KEY,
    name                       TEXT                                                          NOT NULL,
    CREATED_AT                 TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('UTC'::TEXT, NOW()) NOT NULL,
    updated_at                 TIMESTAMP WITH TIME ZONE DEFAULT TIMEZONE('utc'::TEXT, now()) NOT NULL,
    deleted_at                 TIMESTAMP WITH TIME ZONE,
    resource_path              TEXT                     DEFAULT autofillresourcepath(),
    location_type              TEXT,
    partner_internal_id        TEXT,
    partner_internal_parent_id TEXT,
    parent_location_id         TEXT,
    IS_ARCHIVED                BOOLEAN                  DEFAULT FALSE                        NOT NULL,
    access_path                TEXT
);
 
ALTER TABLE IF EXISTS public.locations
    ADD CONSTRAINT fk__locations_parent_location_id
    FOREIGN KEY (parent_location_id)
    REFERENCES public.locations(location_id);

CREATE POLICY rls_locations ON "locations" USING (permission_check(resource_path, 'locations')) WITH CHECK (permission_check(resource_path, 'locations'));
ALTER TABLE "locations"
    ENABLE ROW LEVEL SECURITY;
ALTER TABLE "locations"
    FORCE ROW LEVEL SECURITY;
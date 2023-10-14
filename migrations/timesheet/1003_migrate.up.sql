CREATE TABLE IF NOT EXISTS public.location_types
(
    location_type_id        TEXT                     NOT NULL PRIMARY KEY,
    NAME                    TEXT                     NOT NULL,
    display_name            TEXT,
    parent_name             TEXT,
    parent_location_type_id TEXT,
    updated_at              TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at              TIMESTAMP WITH TIME ZONE NOT NULL,
    deleted_at              TIMESTAMP WITH TIME ZONE,
    resource_path           TEXT    DEFAULT autofillresourcepath(),
    is_archived             BOOLEAN DEFAULT FALSE    NOT NULL,
    CONSTRAINT unique__location_type_name_resource_path UNIQUE (NAME,
                                                                resource_path)
);

ALTER TABLE IF EXISTS public.location_types
    ADD CONSTRAINT fk__location_types__location_type_id
    FOREIGN KEY (location_type_id) REFERENCES location_types(location_type_id);

CREATE policy rls_location_types ON "location_types"
    USING (permission_check(resource_path, 'location_types'))
    WITH CHECK (permission_check(resource_path, 'location_types'));

ALTER TABLE "location_types"
    ENABLE ROW LEVEL SECURITY;

ALTER TABLE "location_types"
    FORCE ROW LEVEL SECURITY;
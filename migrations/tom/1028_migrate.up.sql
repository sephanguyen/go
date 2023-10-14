--- Note that `conversation_locations_locations_fk` is manually removed on staging
--- due to deletion from draft service.
--- See https://manabie.slack.com/archives/C04DHSSA1EG/p1686292296366889
CREATE TABLE IF NOT EXISTS public.conversation_locations (
    "conversation_id" text NOT NULL,
    "location_id" text NOT NULL,
    "access_path" text,
    "created_at" timestamp with time zone NOT NULL,
    "updated_at" timestamp with time zone NOT NULL,
    "deleted_at" timestamp with time zone,
    "resource_path" text DEFAULT autofillresourcepath(),

    CONSTRAINT conversation_locations_pk PRIMARY KEY (conversation_id, location_id),
    CONSTRAINT conversation_locations_conversation_fk FOREIGN KEY (conversation_id) REFERENCES "conversations"(conversation_id),
    CONSTRAINT conversation_locations_locations_fk FOREIGN KEY (location_id) REFERENCES "locations"(location_id)
);

CREATE OR REPLACE function permission_check(resource_path TEXT, table_name TEXT)
RETURNS BOOLEAN 
AS $$
    select ($1 = current_setting('permission.resource_path') )::BOOLEAN
$$  LANGUAGE SQL IMMUTABLE;

CREATE POLICY rls_conversation_locations ON "conversation_locations" using (permission_check(resource_path, 'conversation_locations')) with check (permission_check(resource_path, 'conversation_locations'));

ALTER TABLE "conversation_locations" ENABLE ROW LEVEL security;
ALTER TABLE "conversation_locations" FORCE ROW LEVEL security;

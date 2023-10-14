ALTER TABLE public.info_notifications ADD COLUMN IF NOT EXISTS created_user_id TEXT default NULL;

CREATE TABLE IF NOT EXISTS public.notification_locations (
    notification_id TEXT NOT NULL,
    location_id TEXT NOT NULL,
    access_path  TEXT,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz NOT NULL,
    resource_path TEXT DEFAULT autofillresourcepath(),

    CONSTRAINT notification_location_id_pk PRIMARY KEY (notification_id, location_id)
);

CREATE POLICY rls_notification_locations ON "notification_locations"
USING (permission_check(resource_path, 'notification_locations'))
WITH CHECK (permission_check(resource_path, 'notification_locations'));

ALTER TABLE "notification_locations" ENABLE ROW LEVEL security;
ALTER TABLE "notification_locations" FORCE ROW LEVEL security;

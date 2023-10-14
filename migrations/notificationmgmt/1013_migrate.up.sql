CREATE TABLE IF NOT EXISTS system_notification_contents (
    system_notification_content_id TEXT NOT NULL,
    system_notification_id TEXT NOT NULL,
    "language" TEXT NOT NULL,
    "text" TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT (now() at time zone 'utc'),
    deleted_at TIMESTAMP WITH TIME ZONE,
    resource_path text NOT NULL DEFAULT autofillresourcepath(),

    CONSTRAINT system_notification_contents_system_notification_content_id_pk PRIMARY KEY (system_notification_content_id),
    CONSTRAINT system_notification_contents_system_notification_id_fk FOREIGN KEY (system_notification_id) REFERENCES system_notifications(system_notification_id)
);

CREATE POLICY rls_system_notification_contents ON "system_notification_contents" USING (permission_check(resource_path, 'system_notification_contents')) with check (permission_check(resource_path, 'system_notification_contents'));
CREATE POLICY rls_system_notification_contents_restrictive ON "system_notification_contents" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'system_notification_contents')) WITH CHECK (permission_check(resource_path, 'system_notification_contents'));

ALTER TABLE "system_notification_contents" ENABLE ROW LEVEL security;
ALTER TABLE "system_notification_contents" FORCE ROW LEVEL security;

ALTER TABLE "system_notifications" DROP COLUMN IF EXISTS content;
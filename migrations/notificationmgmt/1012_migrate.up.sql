-- drop all policies
ALTER TABLE important_events DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS rls_important_events_restrictive ON "important_events";
DROP POLICY IF EXISTS rls_important_events ON "important_events";
ALTER TABLE important_event_recipients DISABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS rls_important_event_recipients_restrictive ON "important_event_recipients";
DROP POLICY IF EXISTS rls_important_event_recipients ON "important_event_recipients";

-- drop all constraints
ALTER TABLE important_event_recipients DROP CONSTRAINT IF EXISTS pk__important_event_recipients;
ALTER TABLE important_event_recipients DROP CONSTRAINT IF EXISTS fk__important_event_recipients__important_events;
ALTER TABLE important_events DROP CONSTRAINT IF EXISTS pk__important_events;
ALTER TABLE important_events DROP CONSTRAINT IF EXISTS uk__important_events__reference_id;

-- rename column and table name
ALTER TABLE important_event_recipients RENAME COLUMN important_event_recipient_id TO system_notification_recipient_id;
ALTER TABLE important_event_recipients RENAME COLUMN important_event_id TO system_notification_id;
ALTER TABLE important_event_recipients RENAME TO system_notification_recipients;

ALTER TABLE important_events RENAME COLUMN important_event_id TO system_notification_id;
ALTER TABLE important_events RENAME TO system_notifications;

-- add back constraints
ALTER TABLE system_notifications ADD CONSTRAINT pk__system_notifications PRIMARY KEY (system_notification_id);
ALTER TABLE system_notifications ADD CONSTRAINT uk__system_notifications__reference_id UNIQUE (reference_id);

ALTER TABLE system_notification_recipients
    ADD CONSTRAINT pk__system_notification_recipients PRIMARY KEY (system_notification_recipient_id);
ALTER TABLE system_notification_recipients
    ADD CONSTRAINT fk__system_notification_recipients__system_notifications
    FOREIGN KEY (system_notification_id) REFERENCES system_notifications(system_notification_id);

-- enable rls
CREATE POLICY rls_system_notifications ON "system_notifications" USING (permission_check(resource_path, 'system_notifications')) with check (permission_check(resource_path, 'system_notifications'));
CREATE POLICY rls_system_notifications_restrictive ON "system_notifications" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'system_notifications')) WITH CHECK (permission_check(resource_path, 'system_notifications'));

ALTER TABLE "system_notifications" ENABLE ROW LEVEL security;
ALTER TABLE "system_notifications" FORCE ROW LEVEL security;

CREATE POLICY rls_system_notification_recipients ON "system_notification_recipients" USING (permission_check(resource_path, 'system_notification_recipients')) with check (permission_check(resource_path, 'system_notification_recipients'));
CREATE POLICY rls_system_notification_recipients_restrictive ON "system_notification_recipients" AS RESTRICTIVE TO public USING (permission_check(resource_path, 'system_notification_recipients')) WITH CHECK (permission_check(resource_path, 'system_notification_recipients'));

ALTER TABLE "system_notification_recipients" ENABLE ROW LEVEL security;
ALTER TABLE "system_notification_recipients" FORCE ROW LEVEL security;
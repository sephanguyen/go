CREATE TABLE IF NOT EXISTS public.notification_internal_user (
    user_id TEXT NOT NULL,
    is_system BOOLEAN,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    deleted_at timestamptz,
    resource_path TEXT DEFAULT autofillresourcepath(),
    
    CONSTRAINT notification_internal_user_pk PRIMARY KEY (user_id)
);

CREATE POLICY rls_notification_internal_user ON "notification_internal_user" AS PERMISSIVE
USING (permission_check(resource_path, 'notification_internal_user'))
WITH CHECK (permission_check(resource_path, 'notification_internal_user'));

CREATE POLICY rls_notification_internal_user_restrictive ON "notification_internal_user" AS RESTRICTIVE
USING (permission_check(resource_path, 'notification_internal_user'))
WITH CHECK (permission_check(resource_path, 'notification_internal_user'));

ALTER TABLE "notification_internal_user" ENABLE ROW LEVEL security;
ALTER TABLE "notification_internal_user" FORCE ROW LEVEL security;
